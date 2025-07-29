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

package matrix

import "fmt"

type Matrix struct {
	rows    int
	columns int
	Cells   []int
}

func NewMatrix(rows int, columns int) *Matrix {
	m := Matrix{rows, columns, make([]int, rows*columns)}
	return &m
}

func (m *Matrix) GetOffset(row int, column int) int {
	return m.columns*row + column
}

func (m *Matrix) GetRows() int {
	return m.rows
}

func (m *Matrix) GetColumns() int {
	return m.columns
}

func (m *Matrix) GetCell(row int, column int) int {
	return m.Cells[m.GetOffset(row, column)]
}

func (m *Matrix) SetCell(row int, column int, value int) {
	m.Cells[m.GetOffset(row, column)] = value
}

func (m *Matrix) Reset() {
	offset := 0
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.columns; j++ {
			m.Cells[offset] = 0
			offset++
		}
	}
}

func (m *Matrix) CopyHorizVector(row int, column int, vector []int) {
	offset := m.GetOffset(row, column)
	for _, value := range vector {
		m.Cells[offset] = value
		offset++
	}
}

func (m *Matrix) Dump() {
	offset := 0
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.columns; j++ {
			fmt.Printf("%d ", m.Cells[offset])
			offset++
		}
		fmt.Printf("\n")
	}
}
