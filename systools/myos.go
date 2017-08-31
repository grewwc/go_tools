package systools

import (
	"log"
	"os"
)

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatal(err) //some strange things happens, so I fatal it.
	}
	return true
}
