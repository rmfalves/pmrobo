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

package project

import (
	"goproj/common"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type RootNode struct {
	XMLName   xml.Name      `xml:"project"`
	Calendar  CalendarNode  `xml:"calendar"`
	Resources ResourcesList `xml:"resources"`
	Tasks     TasksList     `xml:"tasks"`
}

type CalendarNode struct {
	XMLName      xml.Name     `xml:"calendar"`
	KickOffDate  string       `xml:"kick-off-date"`
	IdleWeekDays WeekDayList  `xml:"idle-week-days"`
	IdleDates    IdleDateList `xml:"idle-dates"`
}

type WeekDayList struct {
	XMLName     xml.Name `xml:"idle-week-days"`
	IdleWeekDay []string `xml:"idle-week-day"`
}

type IdleDateList struct {
	XMLName  xml.Name `xml:"idle-dates"`
	IdleDate []string `xml:"idle-date"`
}

type ResourcesList struct {
	XMLName  xml.Name       `xml:"resources"`
	Resource []ResourceNode `xml:"resource"`
}

type ResourceNode struct {
	XMLName  xml.Name `xml:"resource"`
	Id       string   `xml:"id,attr"`
	Capacity int      `xml:"capacity,attr"`
}

type TasksList struct {
	XMLName xml.Name   `xml:"tasks"`
	Task    []TaskNode `xml:"task"`
}

type TaskNode struct {
	XMLName          xml.Name         `xml:"task"`
	Id               string           `xml:"id,attr"`
	Duration         int              `xml:"duration"`
	DependenciesList DependenciesList `xml:"dependencies"`
	AllocationsList  AllocationsList  `xml:"allocations"`
}

type DependenciesList struct {
	XMLName    xml.Name         `xml:"dependencies"`
	Dependency []DependencyNode `xml:"dependency"`
}

type DependencyNode struct {
	XMLName         xml.Name `xml:"dependency"`
	DependentTaskId string   `xml:"dependent-task-id,attr"`
	Type            string   `xml:"type,attr"`
}

type AllocationsList struct {
	XMLName    xml.Name         `xml:"allocations"`
	Allocation []AllocationNode `xml:"allocation"`
}

type AllocationNode struct {
	XMLName    xml.Name `xml:"allocation"`
	ResourceId string   `xml:"resource-id,attr"`
	Level      int      `xml:"level,attr"`
}

var WeekDayNameToIndex = map[string]int{"sunday": 0, "monday": 1, "tuesday": 2, "wednesday": 3, "thursday": 4, "friday": 5, "saturday": 6}
var WeekDayNameToIndexAbrev = map[string]int{"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6}

func convertWeekDayNameToIndex(weekDayName string) int {
	index, ok := WeekDayNameToIndex[weekDayName]
	if ok {
		return index
	}
	index, ok = WeekDayNameToIndexAbrev[weekDayName]
	if ok {
		return index
	}
	return -1
}

func (p *Project) importCalendar(xmlTree *RootNode) string {
	err := p.calendar.SetKickOffDate(xmlTree.Calendar.KickOffDate)
	if err != "" {
		return err
	}
	for _, wd := range xmlTree.Calendar.IdleWeekDays.IdleWeekDay {
		i := convertWeekDayNameToIndex(strings.ToLower(wd))
		if i < 0 {
			return fmt.Sprintf("Invalid week day '%s'", wd)
		}
		p.calendar.SetWeekDayStatus(i, false)
	}
	for _, d := range xmlTree.Calendar.IdleDates.IdleDate {
		err := p.calendar.AddIdleDate(d)
		if err != "" {
			return err
		}
	}
	return ""
}

func (p *Project) importResources(xmlTree *RootNode) string {
	for _, r := range xmlTree.Resources.Resource {
		if r.Id == "" || r.Capacity == 0 {
			return fmt.Sprintf("A resource tag is missing one or more attributes")
		}
		err := p.AddResource(r.Id, r.Capacity)
		if err != "" {
			return err
		}
	}
	return ""
}

func (p *Project) importTasks(xmlTree *RootNode) string {
	taskDependencies := map[string]map[string]int{}
	for _, t := range xmlTree.Tasks.Task {
		if t.Id == "" || t.Duration == 0 {
			return fmt.Sprintf("A task tag is missing one or more attributes")
		}
		if t.Duration < 0 {
			return fmt.Sprintf("Task '%s' has zero or negative duration", t.Id)
		}
		err := p.AddTask(t.Id, t.Duration)
		if err != "" {
			return err
		}
		for _, dep := range t.DependenciesList.Dependency {
			depType := common.FS
			if dep.DependentTaskId == "" {
				return fmt.Sprintf("A dependency tag at task '%s' is missing one or more attributes", t.Id)
			}
			if dep.Type != "" {
				depType = common.DepTextToType(dep.Type)
			}
			_, exists := taskDependencies[t.Id]
			if !exists {
				taskDependencies[t.Id] = map[string]int{}
			}
			taskDependencies[t.Id][dep.DependentTaskId] = depType
		}
		for _, alloc := range t.AllocationsList.Allocation {
			if alloc.ResourceId == "" {
				return fmt.Sprintf("An allocation tag at task '%s' is missing one or more attributes", t.Id)
			}
			level := 1
			if alloc.Level != 0 {
				level = alloc.Level
			}
			err := p.AddResourceAllocation(t.Id, alloc.ResourceId, level)
			if err != "" {
				return err
			}
		}
	}
	for task1, m := range taskDependencies {
		for task2, depType := range m {
			err := p.AddTaskDependency(task1, task2, depType)
			if err != "" {
				return err
			}
		}
	}
	return ""
}

func importFromXmlRawBytes(xmlRawBytes []byte) (*Project, string) {
	var xmlTree RootNode
	err := xml.Unmarshal(xmlRawBytes, &xmlTree)
	if err != nil {
		return nil, err.Error()
	}
	p := NewProject()
	errStr := p.importCalendar(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	errStr = p.importResources(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	errStr = p.importTasks(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	return p, ""
}

func ImportFromXmlString(xmlStr string) (*Project, string) {
	return importFromXmlRawBytes([]byte(xmlStr))
}

func ImportFromXmlFile(xmlFileName string) (*Project, string) {
	xmlFile, err := os.Open(xmlFileName)
	defer xmlFile.Close()
	if err != nil {
		return nil, err.Error()
	}
	xmlStr, _ := ioutil.ReadAll(xmlFile)
	return importFromXmlRawBytes([]byte(xmlStr))
}

func ImportFromDirectXMLTree(xmlTree RootNode) (*Project, string) {
	p := NewProject()
	errStr := p.importCalendar(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	errStr = p.importResources(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	errStr = p.importTasks(&xmlTree)
	if errStr != "" {
		return nil, errStr
	}
	return p, ""
}

func (project *Project) ExportToXML(w io.Writer) {
	fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(w, "<project>\n")
	if project.makespan > 0 {
		fmt.Fprintf(w, "%s<makespan>%d</makespan>\n", xmlIndent, project.makespan)
	}
	fmt.Fprintf(w, "%s<resources>\n", xmlIndent)
	for _, r := range project.resources {
		fmt.Fprintf(w, "%s<resource id=\"%s\" capacity=\"%d\"/>\n", strings.Repeat(xmlIndent, 2), r.id, r.capacity)
	}
	fmt.Fprintf(w, "%s</resources>\n", xmlIndent)
	fmt.Fprintf(w, "%s<tasks>\n", xmlIndent)
	for _, t := range project.tasks {
		fmt.Fprintf(w, "%s<task id=\"%s\">\n", strings.Repeat(xmlIndent, 2), t.id)
		fmt.Fprintf(w, "%s<duration>%d</duration>\n", strings.Repeat(xmlIndent, 3), t.duration)
		if t.startT > common.UNDEF {
			fmt.Fprintf(w, "%s<start-t>%d</start-t>\n", strings.Repeat(xmlIndent, 3), t.startT)
		}
		if t.startDate != "" {
			fmt.Fprintf(w, "%s<start-date>%s</start-date>\n", strings.Repeat(xmlIndent, 3), t.startDate)
		}
		if t.finishT > common.UNDEF {
			fmt.Fprintf(w, "%s<finish-t>%d</finish-t>\n", strings.Repeat(xmlIndent, 3), t.finishT)
		}
		if t.finishDate != "" {
			fmt.Fprintf(w, "%s<finish-date>%s</finish-date>\n", strings.Repeat(xmlIndent, 3), t.finishDate)
		}
		if len(t.taskDependencies) == 0 {
			fmt.Fprintf(w, "%s<dependencies/>\n", strings.Repeat(xmlIndent, 3))
		} else {
			fmt.Fprintf(w, "%s<dependencies>\n", strings.Repeat(xmlIndent, 3))
			for depTaskId, depType := range t.taskDependencies {
				fmt.Fprintf(w, "%s<dependency dependent-task-id=\"%s\" type=\"%s\"/>\n", strings.Repeat(xmlIndent, 4), depTaskId, common.DepTypeToText(depType))
			}
			fmt.Fprintf(w, "%s</dependencies>\n", strings.Repeat(xmlIndent, 3))
		}
		if len(t.resourceAllocations) == 0 {
			fmt.Fprintf(w, "%s<allocations/>\n", strings.Repeat(xmlIndent, 3))
		} else {
			fmt.Fprintf(w, "%s<allocations>\n", strings.Repeat(xmlIndent, 3))
			for resourceId, level := range t.resourceAllocations {
				fmt.Fprintf(w, "%s<allocation resource-id=\"%s\" level=\"%d\"/>\n", strings.Repeat(xmlIndent, 4), resourceId, level)
			}
			fmt.Fprintf(w, "%s</allocations>\n", strings.Repeat(xmlIndent, 3))
		}
		fmt.Fprintf(w, "%s</task>\n", strings.Repeat(xmlIndent, 2))
	}
	fmt.Fprintf(w, "%s</tasks>\n", xmlIndent)
	fmt.Fprintf(w, "</project>\n")
}

func (project *Project) ExportScheduleToStringXML() string {
	var w strings.Builder
	fmt.Fprintf(&w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(&w, "<schedule makespan=\"%d\">\n", project.makespan)
	for _, t := range project.tasks {
		fmt.Fprintf(&w, "%s<task id=\"%s\">\n", xmlIndent, t.id)
		fmt.Fprintf(&w, "%s<duration>%d</duration>\n", strings.Repeat(xmlIndent, 2), t.duration)
		if t.startT > common.UNDEF {
			fmt.Fprintf(&w, "%s<start-t>%d</start-t>\n", strings.Repeat(xmlIndent, 2), t.startT)
		}
		if t.startDate != "" {
			fmt.Fprintf(&w, "%s<start-date>%s</start-date>\n", strings.Repeat(xmlIndent, 2), t.startDate)
		}
		if t.finishT > common.UNDEF {
			fmt.Fprintf(&w, "%s<finish-t>%d</finish-t>\n", strings.Repeat(xmlIndent, 2), t.finishT)
		}
		if t.finishDate != "" {
			fmt.Fprintf(&w, "%s<finish-date>%s</finish-date>\n", strings.Repeat(xmlIndent, 2), t.finishDate)
		}
		fmt.Fprintf(&w, "%s</task>\n", xmlIndent)
	}
	fmt.Fprintf(&w, "</schedule>\n")
	return w.String()
}
