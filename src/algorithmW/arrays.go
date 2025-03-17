package algorithmW

func Fill[T any](arr []T, value T) {
	for i := 0; i < len(arr); i++ {
		arr[i] = value
	}
}

func Reverse[T any](arr []T) {
	for i := 0; i < len(arr)/2; i++ {
		arr[i], arr[len(arr)-i-1] = arr[len(arr)-i-1], arr[i]
	}
}
