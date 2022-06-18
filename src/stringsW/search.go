package stringsW

// if target is in slice, return true
// else return false
func Contains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}

	return false
}

func CopySlice(original []string) []string {
	res := make([]string, 0, len(original))
	res = append(res, original...)
	return res
}
