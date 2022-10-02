package _helpers

import "strings"

func GetBucketName(ossKey string) string {
	idx := strings.IndexByte(ossKey, ':')
	if idx < 0 {
		return ossKey
	}
	return ossKey[idx+1:]
}
