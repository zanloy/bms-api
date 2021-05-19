package models_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/stretchr/testify/assert"
	. "github.com/zanloy/bms-api/models"
)

var (
	healthyNode = corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "healthy-node"},
		Spec:       corev1.NodeSpec{},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewQuantity(8, resource.DecimalSI), // 8 core cpu
				corev1.ResourceMemory: *resource.NewQuantity(16, resource.BinarySI), // 16Gi ram
			},
			Phase: corev1.NodeRunning,
			Conditions: []corev1.NodeCondition{{
				Type:   corev1.NodeReady,
				Status: corev1.ConditionTrue,
			}},
			NodeInfo: corev1.NodeSystemInfo{
				KernelVersion:  "1",
				KubeletVersion: "1",
			},
		},
	}

	metrics1m1Mi = metricsv1beta1.NodeMetrics{
		Usage: corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewQuantity(1, resource.DecimalSI),
			corev1.ResourceMemory: *resource.NewQuantity(1024*1024, resource.BinarySI),
		},
	}
)

func TestFromK8Node(t *testing.T) {
	testCases := []struct {
		desc     string
		input    corev1.Node
		expected Node
	}{{
		desc:  "with a healthy node",
		input: healthyNode,
		expected: Node{
			Name:           "healthy-node",
			Healthy:        StatusHealthy,
			Conditions:     []string{"Ready"},
			KernelVersion:  "1",
			KubeletVersion: "1",
			CPU: ResourceQuantities{
				Allocatable: *resource.NewQuantity(8, resource.DecimalSI),
			},
			Memory: ResourceQuantities{
				Allocatable: *resource.NewQuantity(16, resource.BinarySI),
			},
		},
	}}

	for _, testCase := range testCases {
		result := FromK8Node(testCase.input)
		assert.Equal(t, testCase.expected, result, testCase.desc)
	}
}

func TestAddMetrics(t *testing.T) {
	node := FromK8Node(healthyNode)
	node.AddMetrics(metrics1m1Mi)
	assert.Equal(t, "1", node.CPU.Utilized.String())
	assert.Equal(t, "1Mi", node.Memory.Utilized.String())
}
