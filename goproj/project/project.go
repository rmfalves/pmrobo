/****************************************************************************************
GoProj - A lighweight and efficient concurrent SLS solver for the RCPSP problem
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

package project

import (
	"goproj/common"
	"goproj/solver"
	"fmt"
)

const (
	xmlIndent    = "    "
	FIND_OPTIMAL = 0
)

type resource struct {
	id       string
	capacity int
}

type task struct {
	id                  string
	duration            int
	startT              int
	startDate           string
	finishT             int
	finishDate          string
	resourceAllocations map[string]int
	taskDependencies    map[string]int
	earliestStart       int
	earliestFinish      int
	latestStart         int
	latestFinish        int
}

type solverParameters struct {
	maxIterations int
	threads       int
	step          int
	maxTime       int
}

type Project struct {
	tasks       map[string]task
	resources   map[string]resource
	makespan    int
	minMakespan int
	parameters  solverParameters
	calendar    calendar
}

func (t task) SetT(time int) task {
	aux := t
	aux.startT, aux.finishT = time, time+t.duration-1
	return aux
}

func NewProject() *Project {
	param := solverParameters{solver.DEFAULT_MAX_ITERATIONS, solver.DEFAULT_THREADS, solver.DEFAULT_STEP, 0}
	c := NewCalendar()
	p := Project{map[string]task{}, map[string]resource{}, common.UNDEF, common.UNDEF, param, *c}
	return &p
}

func (p *Project) GetMakespan() int {
	return p.makespan
}

func (p *Project) SetMakespan(t int) {
	p.makespan = t
}

func (project *Project) AddResource(id string, capacity int) string {
	if capacity < 0 {
		return fmt.Sprintf("Resource '%s' has negative capacity", id)
	}
	_, duplicate := project.resources[id]
	if duplicate {
		return fmt.Sprintf("Duplicate resource '%s'", id)
	} else {
		project.resources[id] = resource{id, capacity}
		return ""
	}
}

func (project *Project) AddTask(id string, duration int) string {
	_, duplicate := project.tasks[id]
	if duplicate {
		return fmt.Sprintf("Duplicate task '%s'", id)
	} else {
		project.tasks[id] = task{id, duration, common.UNDEF, "", common.UNDEF, "", map[string]int{}, map[string]int{}, common.UNDEF, common.UNDEF, common.UNDEF, common.UNDEF}
		return ""
	}
}

func (project *Project) AddTaskDependency(firstTaskId string, secondTaskId string, dependencyType int) string {
	if dependencyType < common.SS || dependencyType > common.FF {
		return fmt.Sprintf("Illegal dependency type between '%s' and '%s'", firstTaskId, secondTaskId)
	}
	if firstTaskId == secondTaskId {
		return fmt.Sprintf("Dependency between the same task '%s'", firstTaskId)
	}
	_, existsFirst := project.tasks[firstTaskId]
	_, existsSecond := project.tasks[secondTaskId]
	if !existsFirst {
		return fmt.Sprintf("Undefined task '%s'", firstTaskId)
	}
	if !existsSecond {
		return fmt.Sprintf("Undefined task '%s'", secondTaskId)
	}
	_, duplicate1 := project.tasks[firstTaskId].taskDependencies[secondTaskId]
	_, duplicate2 := project.tasks[secondTaskId].taskDependencies[firstTaskId]
	if duplicate1 || duplicate2 {
		return fmt.Sprintf("Dependency already defined between '%s' and '%s'", firstTaskId, secondTaskId)
	}
	project.tasks[firstTaskId].taskDependencies[secondTaskId] = dependencyType
	return ""
}

func (project *Project) AddResourceAllocation(taskId string, resourceId string, level int) string {
	_, existsTask := project.tasks[taskId]
	if !existsTask {
		return fmt.Sprintf("Undefined task '%s'", taskId)
	}
	resource, existsResource := project.resources[resourceId]
	if !existsResource {
		return fmt.Sprintf("Undefined resource '%s'", resourceId)
	}
	if level > resource.capacity {
		return fmt.Sprintf("Resource '%s' allocation for task '%s' exceeds resource capacity", resourceId, taskId)
	}
	if level < 0 {
		return fmt.Sprintf("Resource '%s' has negative allocation for task '%s'", resourceId, taskId)
	}
	_, duplicate := project.tasks[taskId].resourceAllocations[resourceId]
	if duplicate {
		return fmt.Sprintf("Duplicated allocation of resource '%s' to task '%s'", resourceId, taskId)
	}
	project.tasks[taskId].resourceAllocations[resourceId] = level
	return ""
}

func (p *Project) importSchedule(schedule common.TaskSchedule) {
	for id, time := range schedule {
		p.tasks[id] = p.tasks[id].SetT(time)
	}
}

func (p *Project) checkTaskDependencies(t task) string {
	msg := ""
	for depTaskId, depType := range t.taskDependencies {
		if depType == common.FS && p.tasks[depTaskId].startT <= t.finishT {
			msg += fmt.Sprintf("Tasks '%s' and '%s' violate dependency rule %s\n", t.id, depTaskId, common.DepTypeToText(depType))
		}
		if depType == common.SS && p.tasks[depTaskId].startT < t.startT {
			msg += fmt.Sprintf("Tasks '%s' and '%s' violate dependency rule %s\n", t.id, depTaskId, common.DepTypeToText(depType))
		}
		if depType == common.FF && p.tasks[depTaskId].finishT < t.finishT {
			msg += fmt.Sprintf("Tasks '%s' and '%s' violate dependency rule %s\n", t.id, depTaskId, common.DepTypeToText(depType))
		}
		if depType == common.SF && p.tasks[depTaskId].finishT <= t.startT {
			msg += fmt.Sprintf("Tasks '%s' and '%s' violate dependency rule %s\n", t.id, depTaskId, common.DepTypeToText(depType))
		}
	}
	return msg
}

func (p *Project) checkResourceAllocations(r resource) string {
	msg := ""
	demand := make([]int, p.makespan)
	for i := 0; i < p.makespan; i++ {
		demand[i] = 0
	}
	for _, t := range p.tasks {
		if t.startT == common.UNDEF {
			continue
		}
		level, allocated := t.resourceAllocations[r.id]
		if allocated {
			for time := t.startT; time <= t.finishT; time++ {
				demand[time] += level
			}
		}
	}
	for time := 0; time < p.makespan; time++ {
		if demand[time] > r.capacity {
			msg += fmt.Sprintf("Resource '%s' overflows at time=%d (%d > %d)\n", r.id, time, demand[time], r.capacity)
		}
	}
	return msg
}

func (p *Project) CheckScheduleConsistency() string {
	if len(p.tasks) == 0 {
		return ""
	}
	msg := ""
	if p.makespan == common.UNDEF {
		return "Missing makespan, probably empty schedule\n"
	}
	for _, t := range p.tasks {
		if t.startT < 0 || t.finishT < 0 {
			msg += fmt.Sprintf("Task '%s' missing schedule\n", t.id)
		}
		if t.finishT >= p.makespan {
			msg += fmt.Sprintf("Task '%s' overflows project makespan ( %d > %d)\n", t.id, t.finishT, p.makespan)
		}
		msg += p.checkTaskDependencies(t)
	}
	for _, r := range p.resources {
		msg += p.checkResourceAllocations(r)
	}
	return msg
}

func (p *Project) buildConstraintModel() *common.ConstraintModel {
	model := common.NewConstraintModel()
	model.TaskDependencies = map[string]map[string]int{}
	model.ResourceAllocations = map[string]map[string]int{}
	for _, t := range p.tasks {
		model.AddTaskDefinition(t.id, t.duration, t.earliestStart, t.latestStart)
		for taskId, depType := range t.taskDependencies {
			_, exists := model.TaskDependencies[t.id]
			if !exists {
				model.TaskDependencies[t.id] = map[string]int{}
			}
			model.AddTaskDependency(t.id, taskId, depType)
		}
		for resId, level := range t.resourceAllocations {
			_, exists := model.ResourceAllocations[t.id]
			if !exists {
				model.ResourceAllocations[t.id] = map[string]int{}
			}
			model.AddResourceAllocation(t.id, resId, level)
		}
	}
	for _, r := range p.resources {
		model.AddResourceDefinition(r.id, r.capacity)
	}
	model.MinMakespan = p.minMakespan
	return model
}

func (p *Project) convertTimeOffsetsToDate() {
	p.calendar.buildDateMap(p.makespan)
	for id, t := range p.tasks {
		t.startDate = p.calendar.dateMap[t.startT]
		t.finishDate = p.calendar.dateMap[t.finishT]
		p.tasks[id] = t
	}
}

func (p *Project) Schedule(makespan int) bool {
	var res int
	var sched common.TaskSchedule
	p.criticalPath()
	for _, t := range p.tasks {
		// Validate against inconsistent precedence constraints that mess up critical path results
		if t.earliestStart < 0 || t.earliestFinish < 0 {
			return false
		}
	}
	model := p.buildConstraintModel()
	s := solver.NewSolver(*model)
	s.SetParameters(p.parameters.maxIterations, p.parameters.threads, p.parameters.step, p.parameters.maxTime)
	if makespan == FIND_OPTIMAL {
		res, sched = s.SolveOptimalMakespan()
	} else {
		sched = s.SolveFixedMakespan(makespan)
		if sched == nil {
			res = 0
		} else {
			res = makespan
		}
	}
	if res > 0 {
		p.importSchedule(sched)
		p.makespan = res
		p.convertTimeOffsetsToDate()
		return true
	} else {
		return false
	}
}

func (p *Project) GetMinMakespan() int {
	p.criticalPath()
	return p.minMakespan
}

func (p *Project) SetSolverParameters(maxIterations int, threads int, step int, maxTime int) {
	p.parameters = solverParameters{maxIterations, threads, step, maxTime}
}
