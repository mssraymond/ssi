package main

import (
	"cmp"
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

// Setup steps
func setup() (*kubernetes.Clientset, record.EventRecorderLogger) {
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
	return clientset, recorder
}

func main() {
	clientset, recorder := setup()
	// Continuously loop
	for {
		// Throttle for more manageable logging rate
		time.Sleep(5 * time.Second)
		// Map of PodGroups
		podGroupMap := make(map[string]*PodGroup)
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
		// Get all pods
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error fetching pods: %v\n", err)
			continue
		}
		// Build map of PodGroups
		for i := range pods.Items {
			pod := &pods.Items[i]
			podGroupName := cmp.Or(getPodGroupFromAnnotation(pod), pod.Name) // Coalesce
			// PodGroup exists, add to it
			if podGroup, ok := podGroupMap[podGroupName]; ok {
				podGroup.addNodeName(pod)
				podGroup.addPod(pod)
				podGroup.updatePriorityIfGreater(pod)
			} else {
				// Initialize new PodGroup
				pg := PodGroup{
					Id:            podGroupName,
					Pods:          []*v1.Pod{},
					Priority:      getPriority(pod),
					NodeNames:     make(map[string]struct{}),
					SchedulerName: pod.Spec.SchedulerName,
				}
				pg.addNodeName(pod)
				pg.addPod(pod)
				podGroupMap[podGroupName] = &pg
			}
		}
		// Build slice of PodGroups from map
		podGroupSlice := make([]*PodGroup, 0, len(podGroupMap))
		for _, val := range podGroupMap {
			podGroupSlice = append(podGroupSlice, val)
		}
		// Sort, first by Priority, then by # of pods
		sort.Slice(
			podGroupSlice, func(i, j int) bool {
				pi := podGroupSlice[i]
				pj := podGroupSlice[j]
				if pi.Priority != pj.Priority {
					return pi.Priority > pj.Priority
				}
				return len(pi.Pods) > len(pj.Pods)
			},
		)
		// Build set of occupied nodes
		occupiedNodes := make(map[string]bool)
		for _, pg := range podGroupSlice {
			if pg.SchedulerName == "ssi-scheduler" {
				for nodeName := range pg.NodeNames {
					occupiedNodes[nodeName] = true
				}
			}
		}
		// Iterate through PodGroups
		for _, pg := range podGroupSlice {
			// Disregard scheduled PodGroups
			// Any PodGroup with at least 1 associated node is considered already scheduled
			// Failed partial binds require manual pod deletion to reset scheduling
			if len(pg.NodeNames) > 0 {
				continue
			}
			// Disregard PodGroups not using our scheduler
			if pg.SchedulerName != "ssi-scheduler" {
				fmt.Printf("PodGroup %s doesn't have `spec.schedulerName=ssi-scheduler`; disregard\n", pg.Id)
				continue
			}
			// Log current PodGroup and priority
			priorityLabel := getPriorityLabel(pg.Priority)
			fmt.Printf("Scheduling PodGroup `%s` of size %d with priority '%s'...\n", pg.Id, len(pg.Pods), priorityLabel)
			// Per "do not concern yourself with CPU and RAM resource allocation" instruction, pick the first vacant node
			numNodesNeeded := len(pg.Pods) // Aka. # of pods in PodGroup
			targetNodes := make([]string, 0, numNodesNeeded)
			for _, node := range nodes.Items {
				// Acquire nodes NOT in `occupiedNodes` and not "kind-control-plane"
				if !occupiedNodes[node.Name] && node.Name != "kind-control-plane" {
					targetNodes = append(targetNodes, node.Name)
					// Sufficient vacant nodes acquired
					if len(targetNodes) == numNodesNeeded {
						fmt.Printf("Acquired %d vacant node(s) for PodGroup %s\n", len(targetNodes), pg.Id)
						break
					}
				}
			}
			// Not enough vacant nodes, break for fresh statuses
			if len(targetNodes) < numNodesNeeded {
				fmt.Printf("Insufficient vacant nodes available for PodGroup %s\n", pg.Id)
				break
			}
			// Loop through acquired vacant nodes
			for i, targetNode := range targetNodes {
				// Create binding object, matching on slice indices
				targetPod := pg.Pods[i]
				binding := &v1.Binding{
					ObjectMeta: metav1.ObjectMeta{Name: targetPod.Name},
					Target:     v1.ObjectReference{APIVersion: "v1", Kind: "Node", Name: targetNode},
				}
				// Perform bind
				err = clientset.CoreV1().Pods(targetPod.Namespace).Bind(context.TODO(), binding, metav1.CreateOptions{})
				if err == nil { // Successful scheduling of pod on node
					targetPod.Spec.NodeName = targetNode                                                                                                                                    // Update in-memory snapshot of pod for accurate logging
					occupiedNodes[targetNode] = true                                                                                                                                        // Update `occupiedNodes`
					recorder.Eventf(targetPod, v1.EventTypeNormal, "Scheduled", "Successfully scheduled pod %s (namespace %s) on node %s", targetPod.Name, targetPod.Namespace, targetNode) // Events for `kubectl events`
					time.Sleep(1 * time.Second)                                                                                                                                             // Clearly see order in `kubectl events`
				} else { // Failed scheduling of pod on node
					fmt.Printf("Failed to bind pod %s: %v\n", pg.Id, err)
				}
			}
			pg.reportStatus()
		}
	}
}
