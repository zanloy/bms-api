package models // import https://github.com/zanloy/bms-api/models

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

type TenantInfo struct {
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
}

func ParseTenantInfo(namespace string) (tt TenantInfo) {
	tt.Name = "platform"
	tt.Environment = ""

	// See if we can set Tenant/Env from Name
	if strings.Contains(namespace, "-") {
		parts := strings.Split(namespace, "-")
		last := parts[len(parts)-1]
		environments := viper.GetStringSlice("environments")
		fmt.Printf("environments = %v\n", environments)
		sort.Strings(environments)
		idx := sort.SearchStrings(environments, last)
		if idx < len(environments) && environments[idx] == last {
			tt.Environment = last
			tt.Name = strings.Join(parts[:len(parts)-1], "-")
		}
	}

	return
}
