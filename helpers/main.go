package helpers

import (
	"os"
	"sort"
	"strings"
)

func FileExists(filename string) bool {
	filename = os.ExpandEnv(filename)
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ParseTenantAndEnv(name string) (string, string) {
	tenant := "platform"
	env := ""

	// See if we can set Tenant/Env from Name
	if strings.Contains(name, "-") {
		parts := strings.Split(name, "-")
		switch last := parts[len(parts)-1]; last {
		case "cola", "demo", "dev", "int", "ivv", "pat", "pdt", "perf", "preprod", "prod", "prodtest", "sqa", "test", "uat":
			env = last
			tenant = strings.Join(parts[:len(parts)-1], "-")
		}
	}

	return tenant, env
}

func StringInSlice(slice []string, target string) bool {
	// The strings need to be sorted to do a binary search
	sort.Strings(slice)
	i := sort.Search(len(slice), func(i int) bool { return slice[i] >= target })
	return i < len(slice) && slice[i] == target
}
