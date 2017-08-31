package mymath

const o uint32 = 1


type Func interface {
	Value(x float64) float64
}

func Romberg(f Func,a,b float64, n uint32) float64{
	T := make([][]float64, n)
	var c = make(chan float64)
	h := b-a
	assist := func(lo, hi, h float64, c chan float64){
		temp := 0.0
		for i:=lo; i!=hi+1; i++{
			temp += f.Value(a+h*float64(2*i-1))
		}
		c<-temp
	}
	T[0] = make([]float64, 1)
	T[0][0] = h/2*(f.Value(a)+f.Value(b))
	h/=2
	var j, k uint32
	for j = 1; j != n; j++ {
		T[j] = make([]float64, j+1)
		T[j][0] = 0.5*T[j-1][0] + h * func() float64 {
			//return temp
			if j == 1 {
				return f.Value(a + h)
			}
			if j == 2 {
				return f.Value(a+h) + f.Value(a+h*3)
			}
			go assist( 1, float64(o<<(j-1)/4), h, c )
			go assist(float64(o<<(j-1)/4)+1, float64(o<<(j-1)/2),h,c)
			go assist(float64(o<<(j-1)/2)+1, float64(o<<(j-1)/4*3),h,c)
			go assist(float64(o<<(j-1)/4*3)+1, float64(o<<(j-1)),h,c)
			return <-c+ <-c+ <-c+<-c

		}() //Get the T_j1

		for k = 1; k != j+1; k++ {
			T[j][k] = T[j][k-1] + (T[j][k-1]-T[j-1][k-1])/float64(o<<(2*k)-1)
		}
		h /= 2
	}
	return T[n-1][n-1]
}