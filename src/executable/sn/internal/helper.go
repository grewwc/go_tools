package _helper

import "strings"

func GetOssKey(ossKey string) string {
	key := "oss://"
	idx := strings.Index(ossKey, key)
	if idx < 0 {
		return ossKey
	}
	return ossKey[idx+len(key):]
}
