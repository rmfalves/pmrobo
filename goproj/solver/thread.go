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
)

func (s *Solver) exploreVariables(score int, thread int) {
	bestVar := common.UNDEF
	var bestValue int
	bestNewScore := score
	var newScore int
	var bestUpdatedScores []int
	numVariables := len(s.variables)
varLoop:
	for {
		s.mutexVar.Lock()
		v := s.nextVar
		if s.nextVar < numVariables {
			s.nextVar++
		}
		s.mutexVar.Unlock()
		if v == numVariables {
			break varLoop
		}
		if s.variables[v].lbound == s.variables[v].ubound {
			continue
		}
		x0 := s.variables[v].value
		lbound := x0 - s.param.step
		if lbound < s.variables[v].lbound {
			lbound = s.variables[v].lbound
		}
		ubound := x0 + s.param.step
		if ubound > s.variables[v].ubound {
			ubound = s.variables[v].ubound
		}
		for x := lbound; x <= ubound; x++ {
			newScore = score
			s.mutexStop.Lock()
			if s.stop { // Check STOP flag from other threads
				s.mutexStop.Unlock()
				break varLoop
			}
			s.mutexStop.Unlock()
			if x == x0 {
				continue
			}
			updatedScores := []int{}
			for i, c := range s.variables[v].constraints {
				if s.variables[v].constraints[i] >= s.resourcesOffset {
					continue
				}
				eval := s.evaluate(c, v, x)
				newScore += (eval - s.constraints[c].score) * s.constraints[c].weight
				updatedScores = append(updatedScores, c)
				updatedScores = append(updatedScores, eval)
			}
			n := len(s.constraints)
			for i := s.resourcesOffset + x0; i < n; i += s.makespan {
				f := i + s.durations[v] - 1
				for c := i; c <= f; c++ {
					eval := s.evaluate(c, v, x)
					newScore += (eval - s.constraints[c].score) * s.constraints[c].weight
					updatedScores = append(updatedScores, c)
					updatedScores = append(updatedScores, eval)
				}
			}
			for i := s.resourcesOffset + x; i < n; i += s.makespan {
				f := i + s.durations[v] - 1
				for c := i; c <= f; c++ {
					eval := s.evaluate(c, v, x)
					newScore += (eval - s.constraints[c].score) * s.constraints[c].weight
					updatedScores = append(updatedScores, c)
					updatedScores = append(updatedScores, eval)
				}
			}
			updatedScores = append(updatedScores, -1)
			if newScore < bestNewScore {
				bestVar = v
				bestValue = x
				bestNewScore = newScore
				bestUpdatedScores = updatedScores
				if newScore == 0 {
					s.mutexStop.Lock()
					s.stop = true // Raise STOP flag to other threads
					s.mutexStop.Unlock()
					break varLoop
				}
			}
		}
	}
	s.statusChannel <- thread
	s.varChannels[thread] <- bestVar
	if bestVar > common.UNDEF {
		s.varChannels[thread] <- bestValue
		s.varChannels[thread] <- bestNewScore
		for _, updatedScore := range bestUpdatedScores {
			s.varChannels[thread] <- updatedScore
		}
	}
}
