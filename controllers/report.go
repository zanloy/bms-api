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
	if nodes, err := kubernetes.Nodes().List(labels.Everything()); err == nil {
		// Make the nodes list the size of our nodes
		nodelist := make([]models.ReportNode, len(nodes))
		// Loop through the nodes
		for idx, node := range nodes {
			rn := models.ReportNode{Name: node.Name}
			// Conditions
			for _, condition := range node.Status.Conditions {
				if condition.Status == v1.ConditionTrue {
					rn.Conditions = append(rn.Conditions, string(condition.Type))
				}
			}
			// Versions
			rn.KernelVersion = node.Status.NodeInfo.KernelVersion
			rn.KubeletVersion = node.Status.NodeInfo.KubeletVersion
			// CPU Metrics
			rn.CPU = models.ReportResourceData{
				Allocatable: node.Status.Allocatable["cpu"],
				// TODO: add allocated and utilized
			}
			// Memory Metrics
			rn.Memory = models.ReportResourceData{
				Allocatable: node.Status.Allocatable["memory"],
				// TODO: add allocated and utilized
			}
			// Add to array
			nodelist[idx] = rn
			fmt.Printf("rn = %+v\n", rn)
		}
		// Attach nodelist to report
		report.Nodes = nodelist
	} else {
		report.Errors = append(report.Errors, err.Error())
	}

	ctx.JSON(200, report)
}
