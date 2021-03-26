package kubernetes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zanloy/bms-api/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	ogkubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func startK8(stopCh <-chan struct{}) ogkubernetes.Interface {
	clientset := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test1"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test2"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test3"}},
	)

	// setup mock server
	kubernetes.Clientset = clientset

	// Start kubernetes component
	factory := informers.NewSharedInformerFactory(clientset, 0)
	go factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)
	kubernetes.Factory = factory

	return clientset
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
