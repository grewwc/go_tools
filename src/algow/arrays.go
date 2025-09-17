package algow

import (
	"github.com/grewwc/go_tools/src/typesw"
)

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

func BisectLeft[T any](arr []T, target T, cmp typesw.CompareFunc[T]) int {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	lo, hi := 0, len(arr)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if cmp(target, arr[mid]) > 0 {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo

}

func BisectRight[T any](arr []T, target T, cmp typesw.CompareFunc[T]) int {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	lo, hi := 0, len(arr)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if cmp(target, arr[mid]) < 0 {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return lo
}

func LongestIncreasingSubsequence[T any](arr []T, cmp typesw.CompareFunc[T]) []T {
	if len(arr) == 0 {
		return []T{}
	}
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	sub := make([]T, 0)
	for _, num := range arr {
		idx := BisectLeft(sub, num, cmp)
		// fmt.Print(sub, "-->")
		if idx < 0 || idx >= len(sub) {
			sub = append(sub, num)
		} else {
			sub[idx] = num
		}
		// fmt.Println(sub)
	}
	return sub
}

// EditDistance calculates the minimum edit distance between two arrays using dynamic programming.
// It uses the Wagner-Fischer algorithm to compute the Levenshtein distance.
//
// Parameters:
//   - a1: the first array of elements
//   - a2: the second array of elements
//   - cmp: a comparison function that compares two elements of type T
//     Returns negative if first < second, 0 if equal, positive if first > second
//     If nil, a default comparison function will be used
//
// Returns:
//
//	The minimum number of single-element edits (insertions, deletions or substitutions)
//	required to change array a1 into array a2
func EditDistance[T any](a1, a2 []T, cmp typesw.CompareFunc[T]) int {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	m, n := len(a1), len(a2)
	// Optimize space by ensuring a1 is the longer array
	if m < n {
		return EditDistance(a2, a1, cmp)
	}
	// Initialize previous row with base case values (distance from empty string)
	prev := make([]int, n+1)
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	// Process each element of the first array
	for i := 1; i <= m; i++ {
		curr := make([]int, n+1)
		curr[0] = i // Base case: distance from empty string
		// Process each element of the second array
		for j := 1; j <= n; j++ {
			cost := 0
			if cmp(a1[i-1], a2[j-1]) != 0 {
				cost++
			}
			// Calculate costs for three possible operations
			replace := prev[j-1] + cost // Replace operation
			insert := curr[j-1] + 1     // Insert operation
			remove := prev[j] + 1       // Remove operation
			// Choose the operation with minimum cost
			curr[j] = Min(insert, remove, replace)
		}
		prev = curr
	}
	return prev[n]
}

func Equals[T any](a, b []T, cmp typesw.CompareFunc[T]) bool {
	if len(a) != len(b) {
		return false
	}
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	for i := 0; i < len(a); i++ {
		if cmp(a[i], b[i]) != 0 {
			return false
		}
	}
	return true
}
