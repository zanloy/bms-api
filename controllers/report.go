package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/zanloy/bms-api/kubernetes"
	"github.com/zanloy/bms-api/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ReportController struct{}

// Create is an api endpoint that is called by an external entity (usually a
// cron) to create a report and optionally store it in kubernetes.
func (ctl *ReportController) Create(ctx *gin.Context) {
	report := models.NewReport()

	// Nodes
	if k8nodes, err := kubernetes.Nodes().List(labels.Everything()); err == nil {
		// Make the nodes list the size of our nodes
		nodelist := make([]models.Node, len(k8nodes))
		// Loop through the nodes
		for idx, k8node := range k8nodes {
			node := models.FromK8Node(*k8node)
			// Conditions
			for _, condition := range k8node.Status.Conditions {
				if condition.Status == v1.ConditionTrue {
					node.Conditions = append(node.Conditions, string(condition.Type))
				}
			}
			// Versions
			node.KernelVersion = k8node.Status.NodeInfo.KernelVersion
			node.KubeletVersion = k8node.Status.NodeInfo.KubeletVersion
			// CPU Metrics
			if metrics, err := kubernetes.GetNodeMetrics(node); err == nil {
				node.CPU = models.ResourceQuantities{
					Allocatable: k8node.Status.Allocatable["cpu"],
					// TODO: add allocated
					Utilized: metrics.Usage["cpu"],
				}
				// Memory Metrics
				node.Memory = models.ResourceQuantities{
					Allocatable: k8node.Status.Allocatable["memory"],
					// TODO: add allocated
					Utilized: metrics.Usage["memory"],
				}
			}
			// Add to array
			nodelist[idx] = node
			fmt.Printf("node = %+v\n", node)
		}
		// Attach nodelist to report
		report.Nodes = nodelist
	} else {
		report.Errors = append(report.Errors, err.Error())
	}

	ctx.JSON(200, report)
}
