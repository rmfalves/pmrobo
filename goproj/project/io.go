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
	"bufio"
	"goproj/common"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	delta = 5
)

type dependencyPair struct {
	taskId1 string
	taskId2 string
}

func buildResourceId(id int) string {
	return fmt.Sprintf("R%04d", id)
}

func buildTaskId(id int) string {
	return fmt.Sprintf("T%04d", id)
}

func LoadPspLibRcpFile(filename string) (*Project, string) {
	var numTasks, numResources int
	file, err := os.Open(filename)
	if err != nil {
		//fmt.Print(err)
		return nil, "File not found"
	}
	defer file.Close()
	p := NewProject()
	lineNum := 0
	scanner := bufio.NewScanner(file)
	dependencies := []dependencyPair{}
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		r, _ := regexp.Compile("[0-9]+")
		tokens := r.FindAllString(line, -1)
		params := []int{}
		for _, token := range tokens {
			x, _ := strconv.Atoi(token)
			params = append(params, x)
		}
		if lineNum == 1 {
			numTasks = params[0] - 2
			numResources = params[1]
		} else if lineNum == 2 {
			for i := 0; i < numResources; i++ {
				p.AddResource(buildResourceId(i+1), params[i])
			}
		} else if lineNum == 3 {
			continue
		} else if lineNum-3 > numTasks {
			break
		} else {
			taskId := buildTaskId(lineNum - 3)
			p.AddTask(taskId, params[0])
			for res := 1; res <= numResources; res++ {
				if params[res] > 0 {
					p.AddResourceAllocation(taskId, buildResourceId(res), params[res])
				}
			}
			numDependencies := params[numResources+1]
			for i := 0; i < numDependencies; i++ {
				t := params[numResources+2+i] - 1
				if t <= numTasks {
					dependencies = append(dependencies, dependencyPair{taskId, buildTaskId(t)})
				}
			}
		}
	}
	for _, dep := range dependencies {
		p.AddTaskDependency(dep.taskId1, dep.taskId2, common.FS)
	}
	return p, ""
}

func (p *Project) Ganttify(w io.Writer) {
	if len(p.tasks) == 0 {
		fmt.Fprintf(w, "Project has no tasks")
		return
	}
	maxIdLength := 0
	for taskId := range p.tasks {
		if len(taskId) > maxIdLength {
			maxIdLength = len(taskId)
		}
	}
	fmt.Fprintf(w, "%s-|", strings.Repeat("-", maxIdLength))
	for i := 1; i <= p.makespan+1; i++ {
		if i%5 == 0 || i == p.makespan+1 {
			fmt.Fprintf(w, "|")
		} else {
			fmt.Fprintf(w, "-")
		}
	}
	fmt.Fprintf(w, "\n")
	taskIds := []string{}
	for id := range p.tasks {
		taskIds = append(taskIds, id)
	}
	sort.Strings(taskIds)
	for _, id := range taskIds {
		t := p.tasks[id]
		fmt.Fprintf(w, "%s", t.id)
		for i := len(t.id); i < maxIdLength; i++ {
			fmt.Fprintf(w, " ")
		}
		fmt.Fprintf(w, " |")
		if t.startT > common.UNDEF {
			pad1 := strings.Repeat(" ", t.startT)
			bar := strings.Repeat("#", t.finishT-t.startT+1)
			pad2 := strings.Repeat(" ", p.makespan-t.finishT-1)
			fmt.Fprintf(w, "%s%s%s", pad1, bar, pad2)
		} else {
			fmt.Fprintf(w, "%s", strings.Repeat(" ", p.makespan))
		}
		fmt.Fprintf(w, "|\n")
	}
	fmt.Fprintf(w, "%s-|%s|\n", strings.Repeat("-", maxIdLength), strings.Repeat("-", p.makespan))
}
