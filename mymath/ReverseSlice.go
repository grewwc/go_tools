package mymath

//this function just take a slice of int
func Reverse(a *[]int) {
	length := len(*a)
	middle := length / 2
	switch length % 2 {
	case 0:
		for middle >= 0 {
			(*a)[middle], (*a)[length-1-middle] = (*a)[length-1-middle], (*a)[middle]
			middle--
		}
	case 1:
		for middle >= 1 {
			(*a)[middle-1], (*a)[length-middle] = (*a)[length-middle], (*a)[middle-1]
			middle--
		}
	}
}
