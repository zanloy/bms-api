package helpers

import (
	"os"
)

func FileExists(filename string) bool {
	filename = os.ExpandEnv(filename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
