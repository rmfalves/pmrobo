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

package common

type taskDefinition struct {
	Duration      int
	EarliestStart int
	LatestStart   int
}

type ConstraintModel struct {
	TaskDefinitions     map[string]taskDefinition
	ResourceDefinitions map[string]int
	TaskDependencies    map[string]map[string]int
	ResourceAllocations map[string]map[string]int
	MinMakespan         int
}

func NewConstraintModel() *ConstraintModel {
	ConstraintModel := ConstraintModel{map[string]taskDefinition{}, map[string]int{}, map[string]map[string]int{}, map[string]map[string]int{}, 0}
	return &ConstraintModel
}

func (cs *ConstraintModel) AddTaskDefinition(id string, duration int, es int, ls int) {
	cs.TaskDefinitions[id] = taskDefinition{duration, es, ls}
}

func (cs *ConstraintModel) AddResourceDefinition(id string, capacity int) {
	cs.ResourceDefinitions[id] = capacity
}

func (cs *ConstraintModel) AddTaskDependency(taskId1 string, taskId2 string, depType int) {
	cs.TaskDependencies[taskId1][taskId2] = depType
}

func (cs *ConstraintModel) AddResourceAllocation(taskId string, resourceId string, level int) {
	cs.ResourceAllocations[taskId][resourceId] = level
}
