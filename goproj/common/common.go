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

const (
	SS = iota
	SF
	FS
	FF
)

const UNDEF = -1

type TaskSchedule map[string]int

func DepTypeToText(depType int) string {
	switch depType {
	case SS:
		return "SS"
	case SF:
		return "SF"
	case FS:
		return "FS"
	case FF:
		return "FF"
	}
	return ""
}

func DepTextToType(depText string) int {
	switch depText {
	case "SS":
		return SS
	case "SF":
		return SF
	case "FS":
		return FS
	case "FF":
		return FF
	}
	return UNDEF
}
