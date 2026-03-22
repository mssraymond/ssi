package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

// Predefined priority order
var priorityOrder = map[string]int{
	"high":   1,
	"medium": 2,
	"low":    3,
}

// Get priority of pod according to `priorityOrder` map
func getPriority(pod *v1.Pod) int {
	if label, ok := pod.Labels["priority"]; ok {
		if rank, ok := priorityOrder[label]; ok {
			return rank
		}
	}
	return 4
}

func main() {
	// Connect to kind cluster
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Sprintf("Could not find kubeconfig: %v", err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Could not create Kubernetes client: %v", err))
	}
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: clientset.CoreV1().Events(""),
		},
	)
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		v1.EventSource{Component: "ssi-scheduler"},
	)
	fmt.Println("SSI Scheduler starting...")
	for {
		// Throttle for more manageable logging rate
		time.Sleep(5 * time.Second)
		// Get all pods
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error fetching pods: %v\n", err)
			continue
		}
		// Get all nodes
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing nodes: %v\n", err)
			continue
		}
		// Guard against no ready nodes
		if len(nodes.Items) == 0 {
			fmt.Println("No nodes available")
			continue
		}
		// Sort according to priority
		sort.Slice(pods.Items, func(i, j int) bool {
			return getPriority(&pods.Items[i]) < getPriority(&pods.Items[j])
		})
		// Build set of occupied nodes
		occupiedNodes := make(map[string]bool)
		for _, pod := range pods.Items {
			if pod.Spec.NodeName != "" && pod.Spec.SchedulerName == "ssi-scheduler" {
				occupiedNodes[pod.Spec.NodeName] = true
			}
		}
		// Iterate through pods
		for _, pod := range pods.Items {
			// Disregard scheduled pods
			if pod.Spec.NodeName != "" {
				continue
			}
			// Disregard pods not using our scheduler
			if pod.Spec.SchedulerName != "ssi-scheduler" {
				fmt.Printf("Pod %s doesn't have `spec.schedulerName=ssi-scheduler`; disregard\n", pod.Name)
				continue
			}
			// Log current pod and priority
			priorityLabel, ok := pod.Labels["priority"]
			if !ok {
				priorityLabel = "undefined"
			}
			fmt.Printf("Scheduling pod %s with %s priority...\n", pod.Name, priorityLabel)
			// Per "do not concern yourself with CPU and RAM resource allocation" instruction, pick the first vacant node
			// Get first node NOT in `occupiedNodes` and not "kind-control-plane"
			targetNode := ""
			for _, node := range nodes.Items {
				if !occupiedNodes[node.Name] && node.Name != "kind-control-plane" {
					targetNode = node.Name
					break
				}
			}
			// No vacant nodes, break for fresh statuses
			if targetNode == "" {
				fmt.Printf("No vacant nodes available for pod %s\n", pod.Name)
				break
			}
			// Create binding object
			binding := &v1.Binding{
				ObjectMeta: metav1.ObjectMeta{Name: pod.Name},
				Target:     v1.ObjectReference{APIVersion: "v1", Kind: "Node", Name: targetNode},
			}
			// Perform bind
			err = clientset.CoreV1().Pods(pod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
			if err == nil {
				occupiedNodes[targetNode] = true                                                                                                                      // Update `occupiedNodes`
				recorder.Eventf(&pod, v1.EventTypeNormal, "Scheduled", "Successfully assigned pod %s (namespace %s) to node %s", pod.Name, pod.Namespace, targetNode) // Events for `kubectl events`
				fmt.Printf("Successfully placed pod %s on node %s\n", pod.Name, targetNode)
				time.Sleep(1 * time.Second) // Clearly see order in `kubectl events`
			} else {
				fmt.Printf("Failed to bind pod %s: %v\n", pod.Name, err)
			}
		}
	}
}
