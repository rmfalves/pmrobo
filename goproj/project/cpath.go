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
	"fmt"
)

const (
	sourceTaskId = "__source__"
	sinkTaskId   = "__sink__"
)

type cPathNode struct {
	es           int
	ef           int
	ls           int
	lf           int
	pred         []string
	succ         []string
	unmarkedPred int
	unmarkedSucc int
	marked       bool
}

type cPathNetwork map[string]cPathNode

func (p *Project) buildCriticalPathNetwork() cPathNetwork {
	nodes := cPathNetwork{}
	for _, t := range p.tasks {
		nodes[t.id] = cPathNode{-1, -1, -1, -1, []string{}, []string{}, 0, 0, false}
	}
	nodes[sourceTaskId] = cPathNode{-1, -1, -1, -1, []string{}, []string{}, 0, 0, false}
	nodes[sinkTaskId] = cPathNode{-1, -1, -1, -1, []string{}, []string{}, 0, 0, false}
	for _, t := range p.tasks {
		for depTask, depType := range t.taskDependencies {
			if depType != common.FS {
				continue
			}
			aux1 := nodes[t.id]
			aux1.succ = append(aux1.succ, depTask)
			aux1.unmarkedSucc++
			nodes[t.id] = aux1
			aux2 := nodes[depTask]
			aux2.pred = append(aux2.pred, t.id)
			aux2.unmarkedPred++
			nodes[depTask] = aux2
		}
	}
	for id, node := range nodes {
		if id == sourceTaskId || id == sinkTaskId {
			continue
		}
		if len(node.pred) == 0 {
			aux1 := nodes[sourceTaskId]
			aux1.succ = append(aux1.succ, id)
			aux1.unmarkedSucc++
			nodes[sourceTaskId] = aux1

			node.pred = append(node.pred, sourceTaskId)
			node.unmarkedPred = 1 // Predecessor is only the Finish Task
			nodes[id] = node
		}
		if len(node.succ) == 0 {
			node.succ = append(node.succ, sinkTaskId)
			node.unmarkedSucc = 1 // Successor is only the Finish Task
			nodes[id] = node

			aux2 := nodes[sinkTaskId]
			aux2.pred = append(aux2.pred, id)
			aux2.unmarkedPred++
			nodes[sinkTaskId] = aux2
		}
	}
	return nodes
}

func markFromStart(nodes cPathNetwork, id string, es int, ef int) {
	aux1 := nodes[id]
	aux1.es = es
	aux1.ef = ef
	for _, succId := range nodes[id].succ {
		aux2 := nodes[succId]
		aux2.unmarkedPred--
		nodes[succId] = aux2
	}
	aux1.marked = true
	nodes[id] = aux1
}

func maxPredEarliestFinish(nodes cPathNetwork, id string) int {
	max := 0
	for _, pred := range nodes[id].pred {
		if nodes[pred].ef > max {
			max = nodes[pred].ef
		}
	}
	return max
}

func (p *Project) walkFromStart(nodes cPathNetwork) int {
	markFromStart(nodes, sourceTaskId, 0, 0)
	for {
		for id, node := range nodes {
			if id == sourceTaskId || node.unmarkedPred > 0 || node.marked {
				continue
			}
			es := maxPredEarliestFinish(nodes, id)
			if id == sinkTaskId { // If at Finish Task
				return es
			} else {
				ef := es + p.tasks[id].duration
				markFromStart(nodes, id, es, ef)
			}
		}
	}
}

func markFromFinish(nodes cPathNetwork, id string, ls int, lf int) {
	aux1 := nodes[id]
	aux1.ls = ls
	aux1.lf = lf
	for _, predId := range nodes[id].pred {
		aux2 := nodes[predId]
		aux2.unmarkedSucc--
		nodes[predId] = aux2
	}
	aux1.marked = true
	nodes[id] = aux1
}

func minSuccLatestStart(nodes cPathNetwork, id string) int {
	min := -1
	for _, succ := range nodes[id].succ {
		if min == -1 || nodes[succ].ls < min {
			min = nodes[succ].ls
		}
	}
	return min
}

func (p *Project) walkFromFinish(nodes cPathNetwork, makeSpan int) {
	markFromFinish(nodes, sinkTaskId, makeSpan, makeSpan)
	for {
		for id, node := range nodes {
			if id == sinkTaskId || node.unmarkedSucc > 0 || node.marked {
				continue
			}
			lf := minSuccLatestStart(nodes, id)
			if id == sourceTaskId { // If at Start Task
				return
			} else {
				ls := lf - p.tasks[id].duration
				markFromFinish(nodes, id, ls, lf)
			}
		}
	}
}

func (p *Project) criticalPath() {
	nodes := p.buildCriticalPathNetwork()
	p.minMakespan = p.walkFromStart(nodes)
	for id, node := range nodes {
		node.marked = false
		nodes[id] = node
	}
	p.walkFromFinish(nodes, p.minMakespan)
	for id, node := range nodes {
		if id == sourceTaskId || id == sinkTaskId {
			continue
		}
		task := p.tasks[id]
		task.earliestStart = node.es
		task.earliestFinish = node.ef
		task.latestStart = node.ls
		task.latestFinish = node.lf
		p.tasks[id] = task
	}
}

// Just for debugging purposes
func (p *Project) traceCriticalPath() {
	for id, task := range p.tasks {
		if task.earliestStart == task.latestStart {
			fmt.Printf("%s\n", id)
		}
	}
}
