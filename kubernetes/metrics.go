package kubernetes

import (
	"context"

	"github.com/zanloy/bms-api/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetNodeMetrics(node models.Node) (metricsv1beta1.NodeMetrics, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clientset, err := metricsclient.NewForConfig(Config)
	if err != nil {
		return metricsv1beta1.NodeMetrics{}, err
	}

	metricses, err := clientset.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
	return *metricses, err
}
