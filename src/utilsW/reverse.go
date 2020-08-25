package utilsW

func ReverseBytes(arr []byte) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInts(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInt64(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseInt32(arr []int) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseFloat64(arr []float64) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseFloat32(arr []float64) {
	last := len(arr) - 1
	mid := len(arr) / 2
	for i := 0; i < mid; i++ {
		arr[i], arr[last-i] = arr[last-i], arr[i]
	}
}

func ReverseString(s string) string {
	res := []byte(s)
	ReverseBytes(res)
	return string(res)
}
