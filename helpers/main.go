package helpers

import (
	"k8s.io/api/core/v1"
)

// PodIsHealthy will return whether the Pod object is in a "healthy" state.
// A pod is considered healthy if:
//   * The Ready condition is True
//   * The Ready condition is False but reason is 'PodCompleted'
func PodIsHealthy(pod v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady {
			if condition.Status == v1.ConditionTrue {
				return true
			} else {
				if condition.Reason == "PodCompleted" {
					return true
				}
			}
		}
	}
	return false
}

// PodsAreHealthy takes a PodList and validates the health of each pod. It will
// return true only if all the pods are considered healthy.
func PodsAreHealthy(podlist v1.PodList) bool {
	for _, pod := range podlist.Items {
		if PodIsHealthy(pod) == false {
			return false
		}
	}
	return true
}
