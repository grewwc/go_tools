package dummymath

import (
	"fmt"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/grewwc/go_tools/src/algow"
)

const (
	eps = 1e-10
)

type Matrix[T algow.Number] struct {
	M, N int
	data []T
}

func NewMatrix[T algow.Number](M, N int) *Matrix[T] {
	if M <= 0 || N <= 0 {
		panic("invalid shape")
	}
	res := Matrix[T]{
		M:    M,
		N:    N,
		data: make([]T, M*N),
	}
	return &res
}

func (m *Matrix[T]) checkShape(i, j int) {
	if m == nil || i >= m.M || j >= m.N || i < 0 || j < 0 {
		panic("i, j is invalid")
	}
}

func (m *Matrix[T]) Set(i, j int, val T) {
	m.checkShape(i, j)
	m.data[i*m.N+j] = val
}

func (m *Matrix[T]) Get(i, j int) T {
	m.checkShape(i, j)
	return m.data[i*m.N+j]
}

func (m *Matrix[T]) Add(other *Matrix[T]) *Matrix[T] {
	if m.M != other.M || m.N != other.N {
		panic("matrix shape is invalid")
	}
	res := NewMatrix[T](m.M, other.N)
	for i := 0; i < m.M; i++ {
		for j := 0; j < m.N; j++ {
			res.Set(i, j, m.Get(i, j)+other.Get(i, j))
		}
	}
	return res
}

func (m *Matrix[T]) Mul(other *Matrix[T]) *Matrix[T] {
	if m.N != other.M {
		panic("matrix shape is invalid")
	}
	res := NewMatrix[T](m.M, other.N)
	for i := 0; i < res.M; i++ {
		for j := 0; j < res.N; j++ {
			idx := i*res.N + j
			var elem T
			for k := 0; k < m.N; k++ {
				elem += m.data[i*m.N+k] * other.data[k*other.N+j]
			}
			res.data[idx] = elem
		}
	}
	return res
}

func (m *Matrix[T]) IsSquare() bool {
	if m == nil {
		return false
	}
	return m.M == m.N
}

func (m *Matrix[T]) GetRow(row int) []T {
	if row >= m.M || row < 0 {
		panic("invalid shape")
	}
	res := make([]T, 0, m.N)
	for j := 0; j < m.N; j++ {
		res = append(res, m.Get(row, j))
	}
	return res
}

func (m *Matrix[T]) GetCol(col int) []T {
	if col >= m.N || col < 0 {
		panic("invalid shape")
	}
	res := make([]T, 0, m.M)
	for i := 0; i < m.M; i++ {
		res = append(res, m.Get(i, col))
	}
	return res
}

func (m *Matrix[T]) SwapRow(r1, r2 int) {
	m.checkShape(r1, r2)
	for j := 0; j < m.N; j++ {
		val := m.Get(r1, j)
		m.Set(r1, j, m.Get(r2, j))
		m.Set(r1, j, val)
	}
}

func (m *Matrix[T]) T() *Matrix[T] {
	if m == nil {
		panic("invalid shape")
	}
	res := NewMatrix[T](m.N, m.M)
	for i := 0; i < m.M; i++ {
		for j := 0; j < m.N; j++ {
			res.Set(j, i, m.Get(i, j))
		}
	}
	return res
}

func (m *Matrix[T]) Inverse() *Matrix[T] {
	if !m.IsSquare() {
		panic("not square")
	}
	n := m.M

	// Create augmented matrix [A | I]
	aug := NewMatrix[T](n, n*2)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			aug.Set(i, j, m.Get(i, j))
		}
		aug.Set(i, i+n, 1.0)
	}

	for col := 0; col < n; col++ {
		pivotRow := col
		for row := col + 1; row < n; row++ {
			if algow.Abs(aug.Get(row, col)) > algow.Abs(aug.Get(pivotRow, col)) {
				pivotRow = row
			}
		}
		if float64(algow.Abs(aug.Get(pivotRow, col))) < eps {
			log.Println("Matrix is singular (non-invertible)")
			return nil
		}
		// Swap current row with pivot row
		if pivotRow != col {
			aug.SwapRow(col, pivotRow)
		}
		// Scale pivot row to make pivot = 1
		pivotVal := aug.Get(col, col)
		for j := 0; j < 2*n; j++ {
			aug.Set(col, j, T(float64(aug.Get(col, j))/float64(pivotVal)))
		}

		for row := 0; row < n; row++ {
			if row != col {
				factor := aug.Get(row, col)
				for j := 0; j < 2*n; j++ {
					aug.Set(row, j, aug.Get(row, j)-factor*aug.Get(col, j))
				}
			}
		}
	}

	res := NewMatrix[T](m.M, m.N)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			res.Set(i, j, aug.Get(i, j+n))
		}
	}
	return res
}

func (m *Matrix[T]) String() string {
	var builder strings.Builder
	w := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
	for i := 0; i < m.M; i++ {
		sb := strings.Builder{}
		for j := 0; j < m.N; j++ {
			val := m.Get(i, j)
			sb.WriteString(fmt.Sprintf("%v", val))
			if j+1 < m.N {
				sb.WriteString("\t")
			}
		}
		fmt.Fprintln(w, sb.String())
	}
	w.Flush()
	return builder.String()
}
