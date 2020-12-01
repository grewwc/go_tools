package utilsW

import "runtime"

const (
	WINDOWS = iota
	LINUX
	MAC
	OTHER
)

type PlatformType int

func GetPlatform() PlatformType {
	switch runtime.GOOS {
	case "windows":
		return WINDOWS
	case "linux":
		return LINUX
	case "darwin":
		return MAC
	default:
		return OTHER
	}
}
