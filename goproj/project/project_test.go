/****************************************************************************************
PMRobo - A lighweight and efficient multi-threaded project scheduling engine
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
	"goproj/solver"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"
)

const testRcpSuites = false
const testIterateAll = false

func BuildCriticalPathReport(instancesFilename string, reportFilename string) {
	instancesFile, err := os.Open(instancesFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: file '%s' not found\n", instancesFilename)
		return
	}
	defer instancesFile.Close()
	reportFile, err := os.Create(reportFilename)
	defer reportFile.Close()
	scanner := bufio.NewScanner(instancesFile)
	for scanner.Scan() {
		filename := scanner.Text()
		p, err := LoadPspLibRcpFile(filename)
		if err != "" {
			fmt.Fprintf(reportFile, "Exception - %s\n", err)
			fmt.Fprintf(reportFile, "--------------------------------------------------------\n\n")
			continue
		}
		minMakespan := p.GetMinMakespan()
		fmt.Fprintf(reportFile, "%s\t%d\n", filename, minMakespan)
	}
}

func RunRcpSuite(instancesFilename string, reportFilename string) {
	instancesFile, err := os.Open(instancesFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: file '%s' not found\n", instancesFilename)
		return
	}
	defer instancesFile.Close()
	reportFile, err := os.Create(reportFilename)
	defer reportFile.Close()
	scanner := bufio.NewScanner(instancesFile)
	for scanner.Scan() {
		line := scanner.Text()
		r, _ := regexp.Compile("(.*),([0-9]+)$")
		tokens := r.FindStringSubmatch(line)
		filename := tokens[1]
		t, _ := strconv.Atoi(tokens[2])
		fmt.Fprintf(reportFile, "--------------------------------------------------------\n")
		fmt.Fprintf(reportFile, "%s, %d:\n", filename, t)
		p, err := LoadPspLibRcpFile(filename)
		if err != "" {
			fmt.Fprintf(reportFile, "Exception - %s\n", err)
			fmt.Fprintf(reportFile, "--------------------------------------------------------\n\n")
			continue
		}
		for i := t + delta; i >= t-1; i-- {
			fmt.Fprintf(reportFile, "Solving for t=%d... ", i)
			start := time.Now()
			if p.Schedule(i) {
				elapsed := time.Since(start)
				err := p.CheckScheduleConsistency()
				if err == "" {
					fmt.Fprintf(reportFile, "ok (%s)\n", elapsed)
				} else {
					fmt.Fprintf(reportFile, "Error - inconsistent schedule\n")
					break
				}
			} else {
				elapsed := time.Since(start)
				fmt.Fprintf(reportFile, "No schedule found (%s)\n", elapsed)
				break
			}
		}
		fmt.Fprintf(reportFile, "--------------------------------------------------------\n\n")
	}
}

func TestExample(t *testing.T) {
	if true {
		return
	}
	for i := 1; i <= 1000; i++ {
		testname := fmt.Sprintf("Test %d", i)
		t.Run(testname, func(t *testing.T) {
			p := NewProject()
			p.AddResource("R1", 1)
			p.AddResource("R2", 2)
			p.AddResource("R3", 3)
			p.AddResource("R4", 4)
			p.AddTask("T1", 5)
			p.AddTask("T2", 10)
			p.AddTask("T3", 15)
			p.AddTask("T4", 20)
			p.AddTask("T5", 10)
			p.AddTaskDependency("T1", "T2", common.FS)
			p.AddTaskDependency("T2", "T3", common.FS)
			if p.Schedule(50) {
				res := p.CheckScheduleConsistency()
				if res != "" {
					fileGantt, _ := os.Create("project_gantt.txt")
					fileXml, _ := os.Create("project_dump.xml")
					defer func() {
						fileGantt.Close()
						fileXml.Close()
					}()
					p.Ganttify(fileGantt)
					p.ExportToXML(fileXml)
					t.Fatalf("Schedule is inconsistent\n")
				}
			}
		})
	}
}

func TestRcpSuiteCp(t *testing.T) {
	if !testRcpSuites {
		return
	}
	instancesFilename := "instances_j30_cp.csv"
	instancesFile, err := os.Open(instancesFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: file '%s' not found\n", instancesFilename)
		return
	}
	defer instancesFile.Close()
	testName := fmt.Sprintf("Testing CP algorithm for instances from '%s'", instancesFilename)
	t.Run(testName, func(t *testing.T) {
		scanner := bufio.NewScanner(instancesFile)
		for scanner.Scan() {
			line := scanner.Text()
			r, _ := regexp.Compile("(.*),([0-9]+)$")
			tokens := r.FindStringSubmatch(line)
			filename := tokens[1]
			mustHave, _ := strconv.Atoi(tokens[2])
			p, err := LoadPspLibRcpFile(filename)
			if err != "" {
				fmt.Fprintf(os.Stderr, "Exception - %s\n", err)
				continue
			}
			//fmt.Fprintf(os.Stdout, "Solving '%s'...\n", filename)
			makespan := p.GetMinMakespan()
			if makespan != mustHave {
				t.Errorf("Got %d, want %d", makespan, mustHave)
			}
		}
	})
}

func TestRcpSuites(t *testing.T) {
	if !testRcpSuites {
		return
	}
	for i := 30; i <= 60; i += 30 {
		testname := fmt.Sprintf("Testing J%d instances", i)
		t.Run(testname, func(t *testing.T) {
			suitFilename := fmt.Sprintf("instances_j%d_w.csv", i)
			reportFilename := fmt.Sprintf("report_j%d.txt", i)
			RunRcpSuite(suitFilename, reportFilename)
		})
	}
}

func TestAllSequentialTasksFixed(t *testing.T) {
	for n := 1; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			for i := 1; i <= n; i++ {
				proj.AddTask(fmt.Sprintf("T%04d", i), 5)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * 5
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestAllSequentialTasksVar(t *testing.T) {
	for n := 1; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			for i := 1; i <= n; i++ {
				proj.AddTask(fmt.Sprintf("T%04d", i), i)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * (n + 1) / 2
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompetitiveSequentialTasksPlusSlidingTaskFixed(t *testing.T) {
	for n := 1; n <= 20; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, 3)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			id := fmt.Sprintf("T%04d", n+1)
			proj.AddTask(id, 1)
			proj.AddResourceAllocation(id, "R1", 1)
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n*3 + 1
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompetitiveSequentialTasksPlusSlidingTaskVar(t *testing.T) {
	for n := 1; n <= 20; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, i)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			id := fmt.Sprintf("T%04d", n+1)
			proj.AddTask(id, 1)
			proj.AddResourceAllocation(id, "R1", 1)
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n*(n+1)/2 + 1
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestWithFreeTaskFixed(t *testing.T) {
	for n := 5; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, 3)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			id := fmt.Sprintf("T%04d", n+1)
			proj.AddTask(id, 1)
			proj.AddResourceAllocation(id, "R1", 1)
			id = fmt.Sprintf("T%04d", n+2)
			proj.AddTask(id, 8) // This task with no constraints must not stretch project makespan
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations) - %s", n, err)
			}
			expected := n*3 + 1
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestWithFreeTaskVar(t *testing.T) {
	for n := 5; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, i)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			id := fmt.Sprintf("T%04d", n+1)
			proj.AddTask(id, 1)
			proj.AddResourceAllocation(id, "R1", 1)
			id = fmt.Sprintf("T%04d", n+2)
			proj.AddTask(id, 8) // This task with no constraints must not stretch project makespan
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations) - %s", n, err)
			}
			expected := n*(n+1)/2 + 1
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompetitiveTasksFixed(t *testing.T) {
	for n := 1; n <= 20; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, 3)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * 3
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompetitiveTasksVar(t *testing.T) {
	for n := 1; n <= 10; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, i)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			proj.Schedule(FIND_OPTIMAL)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * (n + 1) / 2
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func (p *Project) scheduleAndCompact(makespan int) {
	p.criticalPath()
	model := p.buildConstraintModel()
	s := solver.NewSolver(*model)
	s.SetParameters(0, 0, 0, 0) // Assume default solver parameters
	s.SolveFixedMakespan(makespan)
	compMakespan := s.CompactSchedule()
	compSched := s.ExportSolution()
	p.importSchedule(compSched)
	p.makespan = compMakespan
}

func TestCompactCompetitiveTasksFixed(t *testing.T) {
	for n := 1; n <= 20; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, 3)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			proj.scheduleAndCompact(3 * n * 10)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := 3 * n
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompactCompetitiveTasksVar(t *testing.T) {
	for n := 1; n <= 20; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, i)
				proj.AddResourceAllocation(id, "R1", 1)
			}
			proj.scheduleAndCompact(n * n)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * (n + 1) / 2
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestCompactSequentialTasksFixed(t *testing.T) {
	for n := 1; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, 5)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			for i := 0; i <= n; i++ {
				proj.scheduleAndCompact(5*n + i)
				err := proj.CheckScheduleConsistency()
				if err != "" {
					t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
				}
				expected := 5 * n
				if proj.makespan != expected {
					t.Errorf("Got %d, expected %d", proj.makespan, expected)
				}
			}
		})
	}
}

func TestCompactSequentialTasksVar(t *testing.T) {
	for n := 1; n <= 30; n++ {
		testname := fmt.Sprintf("Testing %d sequential tasks", n)
		t.Run(testname, func(t *testing.T) {
			proj := NewProject()
			proj.AddResource("R1", 1)
			for i := 1; i <= n; i++ {
				id := fmt.Sprintf("T%04d", i)
				proj.AddTask(id, i)
			}
			for i := 1; i <= n-1; i++ {
				proj.AddTaskDependency(fmt.Sprintf("T%04d", i), fmt.Sprintf("T%04d", i+1), common.FS)
			}
			proj.scheduleAndCompact(n * n)
			err := proj.CheckScheduleConsistency()
			if err != "" {
				t.Errorf("Inconsistent schedule for n=%d (make sure you have enough time or iterations)", n)
			}
			expected := n * (n + 1) / 2
			if proj.makespan != expected {
				t.Errorf("Got %d, expected %d", proj.makespan, expected)
			}
		})
	}
}

func TestIterateAll(t *testing.T) {
	if !testIterateAll {
		return
	}
	for i := 1; i <= 5; i++ {
		TestAllSequentialTasksFixed(t)
		TestAllSequentialTasksVar(t)
		TestCompetitiveSequentialTasksPlusSlidingTaskFixed(t)
		TestCompetitiveSequentialTasksPlusSlidingTaskVar(t)
		TestWithFreeTaskFixed(t)
		TestWithFreeTaskVar(t)
		TestCompetitiveTasksFixed(t)
		TestCompetitiveTasksVar(t)
		TestCompactCompetitiveTasksFixed(t)
		TestCompactCompetitiveTasksVar(t)
		TestCompactSequentialTasksFixed(t)
		TestCompactSequentialTasksVar(t)
	}
}
