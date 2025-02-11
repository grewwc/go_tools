package algorithmW

func Fill[T any](arr []T, value T) {
	for i := 0; i < len(arr); i++ {
		arr[i] = value
	}
}
