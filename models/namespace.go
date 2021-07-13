package models

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
)

type Namespace struct {
	Name    string        `json:"name"`
	Tenant  string        `json:"tenant"`
	Env     string        `json:"env,omitempty"`
	Healthy HealthyStatus `json:"healthy"`
	Errors  []string      `json:"errors,omitempty"`

	Deployments []Deployment `json:"deployments"`
}

// Takes in a corev1.Namespace from k8 and builds a Namespace.
func FromK8Namespace(input *corev1.Namespace, factory informers.SharedInformerFactory) Namespace {
	tenant, env := parseTenantAndEnv(input.Name)
	report, err := HealthReportFor(input, factory)
	if err != nil {
		report = HealthReport{
			Healthy: StatusUnknown,
			Errors:  []string{err.Error()},
		}
	}

	ns := Namespace{
		Name:    input.Name,
		Tenant:  tenant,
		Env:     env,
		Healthy: report.Healthy,
		Errors:  report.Errors,
	}

	// TODO: Get bms configmap
	// Setup values from config

	return ns
}

func parseTenantAndEnv(name string) (string, string) {
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
