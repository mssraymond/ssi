# Build a Custom K8s Pod Scheduler from Scratch

## Goals

- Hands-on practice with Golang; newbie to the language
- More intimate programmatic experience with K8s API
- Distributed systems thinking

## Prerequisites

- Install [Docker](https://docs.docker.com/get-started/get-docker/)
- Install [kind](https://kind.sigs.k8s.io/docs/user/quick-start#installation)
- Install [Go](https://go.dev/doc/install)

## Testing

### Run

```sh
./test.sh
```

### Output

```sh
=== Go + Cluster Setup ===
test -f go.mod || go mod init ssi-scheduler
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
go mod tidy
kind create cluster --config kind-config.yaml
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.35.0) 🖼
 ✓ Preparing nodes 📦 📦 📦 📦 📦  
 ✓ Writing configuration 📜 
 ✓ Starting control-plane 🕹️ 
 ✓ Installing CNI 🔌 
 ✓ Installing StorageClass 💾 
 ✓ Joining worker nodes 🚜 
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Not sure what to do next? 😅  Check out https://kind.sigs.k8s.io/docs/user/quick-start/

=== SSI Scheduler Test ===
go build -o bin/ssi-scheduler .
pod/pod-undefined created
pod/pod-none created
pod/pod-low created
pod/pod-high created
pod/pod-medium created
KUBECONFIG=$HOME/.kube/config ./bin/ssi-scheduler
SSI Scheduler starting...
Scheduling pod pod-high with high priority...
Successfully placed pod pod-high on node kind-worker
Scheduling pod pod-medium with medium priority...
Successfully placed pod pod-medium on node kind-worker2
Scheduling pod pod-low with low priority...
Successfully placed pod pod-low on node kind-worker3
Scheduling pod pod-none with undefined priority...
Successfully placed pod pod-none on node kind-worker4
Scheduling pod pod-undefined with undefined priority...
No vacant nodes available for pod pod-undefined
Scheduling pod pod-undefined with undefined priority...
No vacant nodes available for pod pod-undefined

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
pod-high        1/1     Running   0          20s   10.244.1.2   kind-worker    <none>           <none>
pod-low         1/1     Running   0          20s   10.244.3.2   kind-worker3   <none>           <none>
pod-medium      1/1     Running   0          20s   10.244.4.2   kind-worker2   <none>           <none>
pod-none        1/1     Running   0          20s   10.244.2.2   kind-worker4   <none>           <none>
pod-undefined   0/1     Pending   0          20s   <none>       <none>         <none>           <none>

=== Remove pod ===
pod "pod-none" deleted from default namespace
Scheduling pod pod-undefined with undefined priority...
No vacant nodes available for pod pod-undefined
Scheduling pod pod-undefined with undefined priority...
Successfully placed pod pod-undefined on node kind-worker4

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
pod-high        1/1     Running   0          31s   10.244.1.2   kind-worker    <none>           <none>
pod-low         1/1     Running   0          31s   10.244.3.2   kind-worker3   <none>           <none>
pod-medium      1/1     Running   0          31s   10.244.4.2   kind-worker2   <none>           <none>
pod-undefined   1/1     Running   0          31s   10.244.2.3   kind-worker4   <none>           <none>

=== Scheduling Events (verify priority order) ===
26s         Normal   Scheduled   pod/pod-high        Successfully assigned pod pod-high (namespace default) to node kind-worker
25s         Normal   Scheduled   pod/pod-medium      Successfully assigned pod pod-medium (namespace default) to node kind-worker2
24s         Normal   Scheduled   pod/pod-low         Successfully assigned pod pod-low (namespace default) to node kind-worker3
23s         Normal   Scheduled   pod/pod-none        Successfully assigned pod pod-none (namespace default) to node kind-worker4
7s          Normal   Scheduled   pod/pod-undefined   Successfully assigned pod pod-undefined (namespace default) to node kind-worker4
Cleaning up...
kind delete cluster
rm -rf bin/
```

## Additional improvements

- Group-scheduling: Pods with the same pod-group annotation should all get scheduled (or not if unsufficient nodes)
- Retry mechanism: Current retry is inherent to the unending for-loop; a more sophisticated retry algorithm could entail "booting" already-scheduled/already-bound, lower priority pods when higher priority pods are newly launched but no nodes are available
- Performance: Current bottleneck is definitely O(nlogn) sorting of pods by priority; using a combination of Informer (K8s construct for event-driven pod handling) + priority queue (eliminates redundant re-sorts) could enhance scheduling speed, especially as pod count and cluster size grow
