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
	"strconv"
	"time"
)

const (
	daysPerWeek = 7
)

type calendar struct {
	activeWeekDays [daysPerWeek]bool
	idleDates      map[string]bool
	kickOffDate    string
	dateMap        []string
}

func NewCalendar() *calendar {
	var c calendar
	for i := 0; i < daysPerWeek; i++ {
		c.activeWeekDays[i] = true
	}
	c.idleDates = map[string]bool{}
	c.kickOffDate = time.Now().Format("2006-01-02")
	c.dateMap = nil
	return &c
}

func (c *calendar) AddIdleDate(date string) string {
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err.Error()
	}
	c.idleDates[date] = true
	return ""
}

func (c *calendar) SetKickOffDate(date string) string {
	if date == "" {
		c.kickOffDate = time.Now().Format("2006-01-02")
		return ""
	}
	_, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err.Error()
	}
	c.kickOffDate = date
	return ""
}

func (c *calendar) checkNonEmptyWeek() bool {
	for _, isActiveDay := range c.activeWeekDays {
		if isActiveDay {
			return true
		}
	}
	return false
}

func (c *calendar) SetWeekDayStatus(day int, status bool) string {
	if day < 0 || day >= daysPerWeek {
		return "Day out of week range"
	}
	c.activeWeekDays[day] = status
	if !c.checkNonEmptyWeek() {
		c.activeWeekDays[day] = !status
		return "Week requires at least one active day"
	}
	return ""
}

func (c *calendar) buildDateMap(days int) {
	c.dateMap = make([]string, days)
	yyyy, _ := strconv.Atoi(c.kickOffDate[:4])
	mm, _ := strconv.Atoi(c.kickOffDate[5:7])
	dd, _ := strconv.Atoi(c.kickOffDate[8:10])
	date := time.Date(yyyy, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)
	wd := date.Weekday()
	for i := 0; i < days; i++ {
		dateStamp := date.Format("2006-01-02")
		_, isIdleDate := c.idleDates[dateStamp]
		for !c.activeWeekDays[wd] || isIdleDate {
			date = date.AddDate(0, 0, 1)
			dateStamp = date.Format("2006-01-02")
			_, isIdleDate = c.idleDates[dateStamp]
			wd = (wd + 1) % daysPerWeek
		}
		c.dateMap[i] = dateStamp
		date = date.AddDate(0, 0, 1)
		wd = (wd + 1) % daysPerWeek
	}
}
