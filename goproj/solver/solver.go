/****************************************************************************************
PMRobo - A lightweight and efficient multi-threaded project scheduling engine
Copyright (C) 2023  Rui Alves

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
****************************************************************************************/

package solver

import (
	"goproj/common"
	"goproj/matrix"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	NUM_CONSTRAINT_TYPES   = 2
	STOCK_UP               = 1
	STOCK_DOWN             = -1
	DEFAULT_MAX_ITERATIONS = -1
	DEFAULT_MAX_TIME       = -1 // Milliseconds
	DEFAULT_THREADS        = 4
	DEFAULT_STEP           = 10
)

type variable struct {
	value       int
	lbound      int
	ubound      int
	minUbound   int
	constraints []int
}

type dependencyConstraint struct {
	varA    int
	varB    int
	depType int
}

type parameters struct {
	maxIterations int
	threads       int
	step          int
	maxTime       int
}

type Statistics struct {
	BestScore   int
	Iterations  int
	Assignments int
	Restarts    int
	Escapes     int
}

type constraint struct {
	score  int
	weight int
}

type Solver struct {
	durations       []int
	dependencies    []dependencyConstraint
	capacities      []int
	allocations     matrix.Matrix
	varTranslations map[string]int
	variables       []variable
	stocks          matrix.Matrix
	makespan        int
	minMakespan     int
	resourcesOffset int
	constraints     []constraint
	varChannels     []chan int
	statusChannel   chan int
	mutexVar        sync.Mutex
	mutexStop       sync.Mutex
	nextVar         int
	stop            bool
	model           common.ConstraintModel
	param           parameters
	stats           Statistics
}

func NewSolver(model common.ConstraintModel) *Solver {
	var s Solver
	s.param = parameters{DEFAULT_MAX_ITERATIONS, DEFAULT_THREADS, DEFAULT_STEP, DEFAULT_MAX_TIME}
	s.importConstraintModel(model)
	return &s
}

func (s *Solver) SetParameters(maxIterations int, threads int, step int, maxTime int) {
	if maxIterations > 0 {
		s.param.maxIterations = maxIterations
	} else {
		s.param.maxIterations = DEFAULT_MAX_ITERATIONS
	}
	if threads > 0 {
		s.param.threads = threads
	} else {
		s.param.threads = DEFAULT_THREADS
	}
	if step > 0 {
		s.param.step = step
	} else {
		s.param.step = DEFAULT_STEP
	}
	if maxTime > 0 {
		s.param.maxTime = maxTime
	} else {
		s.param.maxTime = DEFAULT_MAX_TIME
	}
}

func (s *Solver) importConstraintModel(model common.ConstraintModel) {
	s.varTranslations = map[string]int{}
	s.variables = make([]variable, len(model.TaskDefinitions))
	s.durations = make([]int, len(model.TaskDefinitions))
	id := 0
	for taskId, task := range model.TaskDefinitions {
		s.varTranslations[taskId] = id
		s.durations[id] = task.Duration
		s.variables[id].lbound = task.EarliestStart
		s.variables[id].minUbound = task.LatestStart
		s.variables[id].ubound = s.variables[id].minUbound // Default value, will vary along the search process
		s.variables[id].constraints = []int{}
		id++
	}
	resourceTranslation := map[string]int{}
	s.capacities = make([]int, len(model.ResourceDefinitions))
	id = 0
	for resourceId, capacity := range model.ResourceDefinitions {
		resourceTranslation[resourceId] = id
		s.capacities[id] = capacity
		id++
	}
	s.dependencies = []dependencyConstraint{}
	for idTask1, dependency := range model.TaskDependencies {
		a := s.varTranslations[idTask1]
		for idTask2, depType := range dependency {
			b := s.varTranslations[idTask2]
			s.dependencies = append(s.dependencies, dependencyConstraint{a, b, depType})
		}
	}
	s.allocations = *matrix.NewMatrix(len(s.variables), len(s.capacities))
	for taskId, allocation := range model.ResourceAllocations {
		t := s.varTranslations[taskId]
		for resourceId, level := range allocation {
			r := resourceTranslation[resourceId]
			s.allocations.SetCell(t, r, level)
		}
	}
	s.minMakespan = model.MinMakespan
	s.model = model
}

func (s *Solver) buildWorkspace(makeSpan int) {
	s.stocks = *matrix.NewMatrix(len(s.capacities), makeSpan)
	for i, value := range s.capacities {
		for j := 0; j < makeSpan; j++ {
			s.stocks.SetCell(i, j, value)
		}
	}
	s.constraints = make([]constraint, len(s.dependencies)+len(s.capacities)*makeSpan)
	for v := range s.variables {
		s.variables[v].constraints = []int{}
	}
	constraintId := 0
	for _, dependency := range s.dependencies {
		a := dependency.varA
		b := dependency.varB
		s.variables[a].constraints = append(s.variables[a].constraints, constraintId)
		s.variables[b].constraints = append(s.variables[b].constraints, constraintId)
		constraintId++
	}
	for r := range s.capacities {
		for t := 0; t < makeSpan; t++ {
			for v := range s.variables {
				if s.allocations.GetCell(v, r) > 0 {
					s.variables[v].constraints = append(s.variables[v].constraints, constraintId)
				}
			}
			constraintId++
		}
	}
	s.resourcesOffset = len(s.dependencies)
	s.makespan = makeSpan
}

func (s *Solver) resetWorkspace() {
	for i := range s.variables {
		s.variables[i].value = common.UNDEF
	}
	for i := range s.capacities {
		for j := 0; j < s.makespan; j++ {
			s.stocks.SetCell(i, j, s.capacities[i])
		}
	}
	for i := range s.constraints {
		s.constraints[i].score = 0
		s.constraints[i].weight = 1
	}
	s.stats = Statistics{common.UNDEF, 0, 0, 0, 0}
}

func (s *Solver) ExportSolution() common.TaskSchedule {
	solution := common.TaskSchedule{}
	for taskId, varId := range s.varTranslations {
		solution[taskId] = s.variables[varId].value
	}
	return solution
}

func (s *Solver) updateStock(varIndex int, startT int, signal int) {
	for resIndex := 0; resIndex < s.allocations.GetColumns(); resIndex++ {
		demand := s.allocations.GetCell(varIndex, resIndex)
		if demand == 0 {
			continue
		}
		pos := s.stocks.GetOffset(resIndex, startT)
		for i := 0; i < s.durations[varIndex]; i++ {
			s.stocks.Cells[pos] += signal * demand
			pos++
		}
	}
}

func (s *Solver) setVariable(varIndex int, value int) {
	prevStartT := s.variables[varIndex].value
	if prevStartT != common.UNDEF {
		s.updateStock(varIndex, prevStartT, STOCK_UP)
	}
	s.variables[varIndex].value = value
	s.updateStock(varIndex, value, STOCK_DOWN)
}

func (s *Solver) getVariableValueForEval(varIndex int, attemptedVarIndex int, attemptedVarValue int) int {
	if varIndex == attemptedVarIndex {
		return attemptedVarValue
	} else {
		return s.variables[varIndex].value
	}
}

func (s *Solver) evalDependency(constrIndex int, attemptedVar int, attemptedValue int) int {
	c := s.dependencies[constrIndex]
	startA := s.getVariableValueForEval(c.varA, attemptedVar, attemptedValue)
	finishA := startA + s.durations[c.varA] - 1
	startB := s.getVariableValueForEval(c.varB, attemptedVar, attemptedValue)
	finishB := startB + s.durations[c.varB] - 1
	switch c.depType {
	case common.SS:
		if startA > startB {
			return startA - startB
		}
	case common.SF:
		if startA >= finishB {
			return startA - finishB + 1
		}
	case common.FS:
		if finishA >= startB {
			return finishA - startB + 1
		}
	case common.FF:
		if finishA > finishB {
			return finishA - finishB
		}
	}
	return 0
}

func (s *Solver) evalResources(constrIndex int, attemptedVar int, attemptedValue int) int {
	offset := constrIndex - s.resourcesOffset
	stock := s.stocks.Cells[offset]
	if attemptedVar > common.UNDEF {
		time := offset % s.makespan
		resourceId := offset / s.makespan
		if time >= s.variables[attemptedVar].value && time < s.variables[attemptedVar].value+s.durations[attemptedVar] {
			stock += s.allocations.GetCell(attemptedVar, resourceId)
		}
		if time >= attemptedValue && time < attemptedValue+s.durations[attemptedVar] {
			stock -= s.allocations.GetCell(attemptedVar, resourceId)
		}
	}
	if stock < 0 {
		return -stock
	} else {
		return 0
	}
}

func (s *Solver) evaluate(constrIndex int, attemptedVar int, attemptedValue int) int {
	var x int
	if constrIndex < s.resourcesOffset {
		x = s.evalDependency(constrIndex, attemptedVar, attemptedValue)
	} else {
		x = s.evalResources(constrIndex, attemptedVar, attemptedValue)
	}
	return x
}

func (s *Solver) incWeights(globalScore *int) {
	for i := range s.constraints {
		if s.constraints[i].score > 0 {
			s.constraints[i].weight++
			*globalScore += s.constraints[i].score
		}
	}
}

func (s *Solver) searchRange(makespan int) bool {
	var iterate bool
	projSlack := makespan - s.minMakespan
	if projSlack < 0 {
		return false
	}
	for varId, v := range s.variables {
		s.variables[varId].ubound = v.minUbound + projSlack
		delta := s.variables[varId].ubound - v.lbound
		s.setVariable(varId, s.variables[varId].lbound+rand.Intn(delta+1))
	}
	score := 0
	for c := range s.constraints {
		eval := s.evaluate(c, common.UNDEF, common.UNDEF)
		s.constraints[c].score = eval
		score += eval
	}
	if score == 0 {
		return true
	}
	s.stop = false
	s.varChannels = make([]chan int, s.param.threads)
	for thread := 0; thread < s.param.threads; thread++ {
		s.varChannels[thread] = make(chan int)
	}
	s.statusChannel = make(chan int)
	if s.param.maxTime > 0 {
		iterate = true
		time.AfterFunc(time.Millisecond*time.Duration(s.param.maxTime), func() {
			iterate = false
		})
	} else {
		iterate = false
	}
	for tries := 0; tries < s.param.maxIterations || iterate; tries++ {
		bestVar := common.UNDEF
		var bestValue int
		var tmpConstraintScores []int
		s.stats.Iterations++
		s.nextVar = 0
		for thread := 0; thread < s.param.threads; thread++ {
			go s.exploreVariables(score, thread)
		}
		for i := 0; i < s.param.threads; i++ {
			thread := <-s.statusChannel
			threadBestVar := <-s.varChannels[thread]
			if threadBestVar > common.UNDEF {
				threadBestValue := <-s.varChannels[thread]
				bestNewScore := <-s.varChannels[thread]
				if bestNewScore < score {
					bestVar = threadBestVar
					bestValue = threadBestValue
					score = bestNewScore
					tmpConstraintScores = []int{}
					token := <-s.varChannels[thread]
					for token > -1 {
						tmpConstraintScores = append(tmpConstraintScores, token)
						token = <-s.varChannels[thread]
					}
				} else {
					// Just clean the channel from those useless updated scores
					token := <-s.varChannels[thread]
					for token > -1 {
						token = <-s.varChannels[thread]
					}
				}
			}
		}
		if bestVar > common.UNDEF {
			s.setVariable(bestVar, bestValue)
			for j := 0; j < len(tmpConstraintScores); j += 2 {
				c := tmpConstraintScores[j]
				eval := tmpConstraintScores[j+1]
				s.constraints[c].score = eval
			}
			if score == 0 {
				break
			}
		} else {
			s.incWeights(&score)
		}
	}
	for thread := 0; thread < s.param.threads; thread++ {
		close(s.varChannels[thread])
	}
	close(s.statusChannel)
	return score == 0
}

func (s *Solver) sumTasksDurations() int {
	sum := 0
	for _, d := range s.durations {
		sum += d
	}
	return sum
}

func (s *Solver) CompactSchedule() int {
	busy := make([]bool, s.makespan)
	for t := range busy {
		busy[t] = false
	}
	for i, v := range s.variables {
		for t := 0; t < s.durations[i]; t++ {
			busy[v.value+t] = true
		}
	}
	maxStart := -1
	maxFinish := -1
	for i, v := range s.variables {
		if v.value > maxStart {
			maxStart = v.value
		}
		if v.value+s.durations[i] > maxFinish {
			maxFinish = v.value + s.durations[i]
		}
	}
	totalDelta := 0
	delta := 0
	for t := maxStart; t >= 0; t-- {
		if busy[t] {
			if delta > 0 {
				for i, v := range s.variables {
					if v.value > t {
						s.variables[i].value -= delta
					}
				}
				totalDelta += delta
				delta = 0
			}
		} else {
			delta++
		}
	}
	if delta > 0 {
		for i := range s.variables {
			s.variables[i].value -= delta
		}
		totalDelta += delta
	}
	return maxFinish - totalDelta
}

func (s *Solver) SolveFixedMakespan(makespan int) common.TaskSchedule {
	s.buildWorkspace(makespan)
	s.resetWorkspace()
	ok := s.searchRange(makespan)
	if ok {
		return s.ExportSolution()
	} else {
		return nil
	}
}

func (s *Solver) SolveOptimalMakespan() (int, common.TaskSchedule) {
	sched := s.SolveFixedMakespan(s.minMakespan)
	if sched != nil {
		return s.minMakespan, sched
	}
	lBound := s.minMakespan - 1
	uBound := s.sumTasksDurations()
	bestMakespan := uBound
	bestSchedule := s.SolveFixedMakespan(uBound)
	if bestSchedule == nil {
		// The problem is impossible, there are likely constraints inconsistencies
		return common.UNDEF, nil
	}
	for uBound-lBound > 1 {
		makespan := (lBound + uBound) / 2
		sched := s.SolveFixedMakespan(makespan)
		if sched == nil {
			lBound = makespan
		} else {
			bestMakespan = makespan
			bestSchedule = sched
			uBound = makespan
			compMakespan := s.CompactSchedule()
			if compMakespan < makespan {
				bestMakespan = compMakespan
				bestSchedule = s.ExportSolution()
				uBound = compMakespan
			}
		}
		// Completely reset the solver object for the next iteration
		p := s.param
		s = NewSolver(s.model)
		s.SetParameters(p.maxIterations, p.threads, p.step, p.maxTime)
	}
	return bestMakespan, bestSchedule
}

func (s *Solver) Solve() (int, common.TaskSchedule) {
	// An alias for solving optimal schedule
	return s.SolveOptimalMakespan()
}

func (s *Solver) ReportStats() string {
	return fmt.Sprintf("%+v\n", s.stats)
}
