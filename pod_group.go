package main

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type PodGroup struct {
	Id            string
	Pods          []*v1.Pod
	Priority      int // Highest among the group
	NodeNames     map[string]struct{}
	SchedulerName string
}

// Predefined priority order
var priorityOrder = map[string]int{
	"high":   3,
	"medium": 2,
	"low":    1,
}

// Convert priority number to label
func getPriorityLabel(priority int) string {
	for label, order := range priorityOrder {
		if order == priority {
			return label
		}
	}
	return "none"
}

// Get priority of pod according to `priorityOrder` map
func getPriority(pod *v1.Pod) int {
	if label, ok := pod.Labels["priority"]; ok {
		if rank, ok := priorityOrder[label]; ok {
			return rank
		}
	}
	return 0
}

// Get pod-group from pod annotation, or empty string if no such annotation
func getPodGroupFromAnnotation(pod *v1.Pod) string {
	if annotation, ok := pod.Annotations["ssi.scheduler/pod-group"]; ok {
		return annotation
	}
	return ""
}

// Update PodGroup's Priority if passed in pod has higher priority
func (pg *PodGroup) updatePriorityIfGreater(pod *v1.Pod) {
	newPriority := getPriority(pod)
	if newPriority > pg.Priority {
		pg.Priority = newPriority
	}
}

// Add to PodGroup's Pods slice
func (pg *PodGroup) addPod(pod *v1.Pod) {
	pg.Pods = append(pg.Pods, pod)
}

// Add to PodGroup's NodeNames set
func (pg *PodGroup) addNodeName(pod *v1.Pod) {
	if pod.Spec.NodeName != "" {
		pg.NodeNames[pod.Spec.NodeName] = struct{}{}
	}
}

// Report status of all pods in PodGroup
func (pg *PodGroup) reportStatus() {
	fmt.Println("\n*** PodGroup Status Summary ***")
	fmt.Printf("PodGroup ID: %s\n", pg.Id)
	fmt.Printf("PodGroup Size: %d\n", len(pg.Pods))
	fmt.Printf("PodGroup Priority: %s\n", getPriorityLabel(pg.Priority))
	fmt.Println("Pod Details:")
	for _, pod := range pg.Pods {
		fmt.Printf("  Pod %s (in namespace %s) is scheduled on node %s\n", pod.Name, pod.Namespace, pod.Spec.NodeName)
	}
	fmt.Print("\n\n")
}
