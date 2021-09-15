package models // import https://github.com/zanloy/bms-api/models

import (
	"strings"
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
		switch last := parts[len(parts)-1]; last {
		case "cola", "demo", "dev", "int", "ivv", "pat", "pdt", "perf", "preprod", "prod", "prodtest", "sqa", "test", "uat":
			tt.Environment = last
			tt.Name = strings.Join(parts[:len(parts)-1], "-")
		}
	}

	return
}
