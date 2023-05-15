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

package main

import (
	"goproj/project"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	configFile  = "pmrobo.xml"
	defaultPort = 9100
)

type Config struct {
	XMLName xml.Name `xml:"config"`
	Threads int      `xml:"threads"`
	Times   TimeList `xml:"times"`
	Step    int      `xml:"step"`
	Port    int      `xml:"port"`
}

type TimeList struct {
	XMLName xml.Name `xml:"times"`
	Time    []int    `xml:"time"`
}

func LoadConfig(filename string) (*Config, string) {
	var settings Config
	xmlFile, err := os.Open(filename)
	defer xmlFile.Close()
	if err != nil {
		return nil, err.Error()
	}
	xmlStr, _ := ioutil.ReadAll(xmlFile)
	err = xml.Unmarshal([]byte(xmlStr), &settings)
	if err != nil {
		return nil, err.Error()
	}
	if settings.Port == 0 {
		settings.Port = defaultPort
	}
	return &settings, ""
}

func main() {
	config, err := LoadConfig(configFile)
	if config.Threads == 0 {
		fmt.Printf("Running on default number of threads.\n")
	} else {
		fmt.Printf("Running on %d threads.\n", config.Threads)
	}
	if err != "" {
		fmt.Fprint(os.Stderr, err)
		return
	}
	r := gin.Default()
	r.POST("/schedule", func(c *gin.Context) {
		var p project.RootNode
		c.Header("Access-Control-Allow-Origin", "*")
		err := c.BindXML(&p)
		if err == nil {
			proj, errStr := project.ImportFromDirectXMLTree(p)
			if errStr == "" {
				solution := false
				for _, maxTime := range config.Times.Time {
					proj.SetSolverParameters(0, config.Threads, config.Step, maxTime)
					fmt.Printf("Trying with time=%d\n", maxTime)
					if proj.Schedule(project.FIND_OPTIMAL) {
						errStr = proj.CheckScheduleConsistency()
						if errStr == "" {
							c.Header("Content-Type", "application/xml")
							c.String(http.StatusOK, proj.ExportScheduleToStringXML())
						} else {
							c.String(http.StatusBadRequest, "Reserved error.")
						}
						solution = true
						break
					}
				}
				if !solution {
					c.String(http.StatusBadRequest, "No schedule found. Project constraints may be too complex or even inconsistent.")
				}
			} else {
				c.String(http.StatusBadRequest, errStr)
			}
		} else {
			c.String(http.StatusBadRequest, err.Error())
		}
	})
	r.Run(fmt.Sprintf(":%d", config.Port))
}
