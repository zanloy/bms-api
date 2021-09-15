package kubernetes_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zanloy/bms-api/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func genNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
	}
}

func startK8(stopCh <-chan struct{}) {
	// setup mock server
	kubernetes.Clientset = fake.NewSimpleClientset(
		genNamespace("test1"),
		genNamespace("test2"),
		genNamespace("test3"),
	)

	if namespaces, err := kubernetes.Clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{}); err == nil {
		fmt.Printf("namespaces = %+v\n", namespaces)
	} else {
		fmt.Printf("There was an error trying to get namespace list: %s\n", err.Error())
	}

	// Start kubernetes component
	kubernetes.Start(stopCh)
}

func TestNamespaceExists(t *testing.T) {
	cases := []struct {
		desc      string
		namespace string
		expected  bool
	}{{
		desc:      "where namespace exists",
		namespace: "test2",
		expected:  true,
	}, {
		desc:      "where namespace doesn't exist",
		namespace: "thisIsNotTheNamespaceYouAreLookingFor",
		expected:  false,
	}}

	stopCh := make(chan struct{})
	defer close(stopCh)
	startK8(stopCh)

	for _, testcase := range cases {
		result := kubernetes.NamespaceExists(testcase.namespace)
		assert.Equal(t, testcase.expected, result, testcase.desc)
	}
}

func TestNamespacesArray(t *testing.T) {
	// Setup test cases
	cases := []struct {
		desc     string
		expected []string
		err      bool
	}{{
		desc:     "looking for valid response",
		expected: []string{"test1", "test2", "test3"},
	}}

	// Setup Kubernetes
	stopCh := make(chan struct{})
	defer close(stopCh)
	startK8(stopCh)

	// Loop
	for _, testcase := range cases {
		result, err := kubernetes.NamespacesArray()
		if testcase.err {
			assert.Error(t, err, testcase.desc)
		} else {
			assert.NoError(t, err, testcase.desc)
		}
		assert.Equal(t, testcase.expected, result, testcase.desc)
	}
}
