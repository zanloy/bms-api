package models // import https://github.com/zanloy/bms-api/models

import (
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
		sort.Strings(environments)
		if sort.SearchStrings(environments, last) < len(environments) {
			tt.Environment = last
			tt.Name = strings.Join(parts[:len(parts)-1], "-")
		}
	}

	return
}
