package machineLearning

import (
	"fmt"
	"log"
	"math"
	"mymath"
)

type LinearRegression struct {
	Learned    bool
	Method     string
	Parameters []float64
	Initial    []float64
	Alpha      float64
}

//Linear_Regression takes 4/5 parameters.
//1.  x ---- 2-d slice, representing "feature"
//2.  y ---- 1-d slice, representing "target"
//3.  initial ---- initial value for intrested parameters, constant is the first one
//4.  alpha ---- set up how long you want to "walk" for 1 step
//5.  iteration ---- (1) maximux iteration, which is used under conditions that without convergence
//	 		    ---- (2) method: including "Batch", "Stochatic"
func (f *LinearRegression) Learn(x [][]float64, y []float64, iteration interface{}) []float64 {
	if len(x) != len(y) {
		log.Fatal("dimensions of x, y NOT the same!")
		return nil
	}

	var count int64
	var recursion int64

	if iteration == nil {
		recursion = 1000
	} else {
		switch t := iteration.(type) {
		case int:
			recursion = int64(t)
		default:
			log.Println("wrong type")
		}
	}

	if f.Method == "" {
		f.Method = "batch"
	} //default method
	m := len(x)
	n := len(x[0])
	var initial []float64
	if len(f.Initial) == 0 {
		initial = make([]float64, n+1)
	} else {
		initial = f.Initial
	}
	var alpha float64
	if f.Alpha == 0.0 {
		alpha = 1e-4
	} else {
		alpha = f.Alpha
	}

	h := func(xj []float64) float64 {
		sum := 0.0
		for i := 0; i != n; i++ {
			sum += initial[i+1] * xj[i]
		}
		return sum + initial[0]
	}

	c := make(chan float64)
	var calc_next func([]float64) []float64
	if f.Method == "batch" {
		calc_next = func(initial_inner []float64) []float64 {
			next := make([]float64, n+1)
			for i := 1; i != n+1; i++ {
				go func(c chan float64) {
					sum := 0.0
					for j := 0; j != m; j++ {
						sum += (h(x[j]) - y[j]) * x[j][i-1]
					}
					c <- sum
				}(c)
				assist := <-c // * math.Exp(-assist/20)
				next[i] = initial_inner[i] - alpha*assist
			}
			temp := 0.0
			for j := 0; j != m; j++ {
				temp += h(x[0]) - y[0]
			}
			next[0] = initial_inner[0] - alpha*temp
			return next
		}
	} else {
		calc_next = func(thetaI []float64) []float64 {
			for i := 1; i != n+1; i++ {
				sum := 0.0
				for j := 0; j != m; j++ {
					sum += (h(x[j]) - y[j]) * x[j][i-1]
				}
				thetaI[i] -= alpha * sum
			}
			sum := 0.0
			for j := 0; j != m; j++ {
				sum += h(x[j]) - y[j]
			}
			thetaI[0] -= alpha * sum
			return thetaI
		}

	}
	initial_matrix := &mymath.Matrix{1, n + 1, [][]float64{initial}}
	dif := initial_matrix
	res := make([]float64, n+1)
	for {
		//fmt.Println(initial)
		next := calc_next(initial)
		initial_matrix.Element[0] = initial
		next_matrix := &mymath.Matrix{1, n + 1, [][]float64{next}}
		dif = next_matrix.Minus(initial_matrix)
		temp := math.Sqrt(dif.Times(dif.T()).Element[0][0])
		if temp <= 1e-10 || count > recursion {
			res = next
			break
		}
		count++
		//println(calc_next(initial)[0])
		initial = next
	}
	f.Learned = true
	f.Parameters = res
	return res
}

func (f LinearRegression) Predict(x []float64) float64 {
	sum := 0.0
	for i, v := range x {
		sum += f.Parameters[i+1] * v
	}
	return sum + f.Parameters[0]
}

func h(theta, x []float64) float64 {
	sum := 0.0
	for i, v := range x {
		sum += theta[i+1] * v
	}
	sum += theta[0]
	return 1.0 / (1 + math.Exp(-sum))
}

type Logistic struct {
	Learned    bool
	Parameters []float64
}

func (f *Logistic) Learn(x [][]float64, y []float64, iteration int) {
	alpha := 0.0
	m := len(y)
	n := len(x[0])
	c := make(chan float64)
	initial := make([]float64, n+1)
	for i, _ := range initial {
		initial[i] = 1
	}
	initial[2] = 1
	var next = make([]float64, n+1)
	copy(next, initial)
	calcNext := func() {
		assist := 0.0
		go func(chan float64) {
			sum := 0.0
			for j := 0; j != m; j++ {
				sum += h(initial, x[j]) - y[j]
			}
			c <- sum
		}(c)
		for i := 0; i != n; i++ {
			for j := 0; j < m; j++ {
				assist = (h(initial, x[j]) - y[j]) * x[j][i]
				alpha = 1e-5 * math.Exp(-assist/20)
				next[i+1] = initial[i+1] - alpha*assist
			}
		}
		temp := <-c
		alpha = 1e-5 * math.Exp(-temp/20)
		next[0] = initial[0] - alpha*temp
	}

	var next_matrix, initial_matrix, diff *mymath.Matrix
	count := 0
	for {
		calcNext()
		next_matrix = &mymath.Matrix{1, n + 1, [][]float64{next}}
		initial_matrix = &mymath.Matrix{1, n + 1, [][]float64{initial}}
		diff = next_matrix.Minus(initial_matrix)
		if math.Sqrt(diff.Times(diff.T()).Element[0][0]) < 1e-5 || count > iteration {
			break
		}
		copy(initial, next)
		count++
	}
	f.Learned = true
	f.Parameters = next
}

func (f *Logistic) Predict(x []float64) int {
	if !f.Learned {
		fmt.Println("not learned yet!")
		return -1
	}

	res := h(f.Parameters, x)

	if res >= 0.5 {
		return 1
	}
	return 0
}
