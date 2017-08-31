package mymath

import (
	"fmt"
	"log"
)

//matrix type
type Matrix struct {
	Dimen_x, Dimen_y int
	Element          [][]float64
}

//creating a new matrix, without this function is OK
func (m *Matrix) NewMatrix(data [][]float64) {
	m.Dimen_x = len(data)
	m.Dimen_y = len(data[0])
	m.Element = make([][]float64, m.Dimen_x)
	for i := 0; i != m.Dimen_x; i++ {
		m.Element[i] = make([]float64, m.Dimen_y)
		for j := 0; j != m.Dimen_y; j++ {
			m.Element[i][j] = data[i][j]
		}
	}
}

//add 2 matrices
func (m Matrix) Add(others interface{}) *Matrix {
	temp := make([][]float64, m.Dimen_x)
	switch other := others.(type) {
	case float64:
		for i, vx := range m.Element {
			temp[i] = make([]float64, m.Dimen_y)
			for j, _ := range vx {
				temp[i][j] = m.Element[i][j] + other
			}
		}

	case *Matrix:
		for i, vx := range m.Element {
			temp[i] = make([]float64, m.Dimen_y)
			for j, _ := range vx {
				temp[i][j] = m.Element[i][j] + other.Element[i][j]
			}
		}
	}
	return &Matrix{m.Dimen_x, m.Dimen_y, temp}
}

//A-B
func (m Matrix) Minus(other interface{}) *Matrix {
	switch cast := other.(type) {
	case *Matrix:
		return m.Add(cast.Times(-1.0))
	case float64:
		return m.Add(cast * -1)
	}
	return nil

}

//A*B
func (m Matrix) Times(others interface{}) *Matrix {
	res := make([][]float64, m.Dimen_x)
	switch other := others.(type) {
	case float64:
		for i := 0; i != m.Dimen_x; i++ {
			res[i] = make([]float64, m.Dimen_y)
			for j := 0; j != m.Dimen_y; j++ {
				res[i][j] = other * m.Element[i][j]
			}
		}
		return &Matrix{m.Dimen_x, m.Dimen_y, res}
	case *Matrix:
		other_dx := len(other.Element)
		if m.Dimen_y != other_dx {
			log.Fatal("cannot multiply the 2 Matrix!")
			return nil
		}

		for i := 0; i != m.Dimen_x; i++ {
			res[i] = make([]float64, m.Dimen_y)
			for j := 0; j != other.Dimen_y; j++ {
				for k := 0; k != m.Dimen_y; k++ {
					res[i][j] += m.Element[i][k] * other.Element[k][j]
				}
			}
		}
	}
	return &Matrix{m.Dimen_x, m.Dimen_x, res}
}

//Show a matrix
func (m Matrix) Show(n ...int) {
	length := len(n)
	if length == 0 {
		length = 2
	} //set the default value
	for i := 0; i != m.Dimen_x; i++ {
		for j := 0; j != m.Dimen_y; j++ {
			fmt.Printf("%.[2]*[1]f  ", m.Element[i][j], length)
		}
	}
}

//transfer
func (m Matrix) T() *Matrix {
	transfered := make([][]float64, m.Dimen_y)
	for i := 0; i != m.Dimen_y; i++ {
		transfered[i] = make([]float64, m.Dimen_x)
		for j := 0; j != m.Dimen_x; j++ {
			transfered[i][j] = m.Element[j][i]
		}
	}
	return &Matrix{m.Dimen_y, m.Dimen_x, transfered}
}
