package helpers

import (
	"os"
	"sort"
)

func FileExists(filename string) bool {
	filename = os.ExpandEnv(filename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func StringInSlice(slice []string, target string) bool {
	// The strings need to be sorted to do a binary search
	sort.Strings(slice)
	i := sort.Search(len(slice), func(i int) bool { return slice[i] >= target })
	return i < len(slice) && slice[i] == target
}
