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

Thanks for using kind! 😊

=== SSI Scheduler Test ===
go build -o bin/ssi-scheduler .
pod/pod-undefined created
pod/pod-none created
pod/crewmate-1 created
pod/crewmate-2 created
pod/crewmate-3 created
pod/pod-high created
pod/teammate-1 created
pod/teammate-2 created
pod/teammate-3 created
pod/teammate-4 created
pod/gangster-1 created
pod/gangster-2 created
pod/gangster-3 created
KUBECONFIG=$HOME/.kube/config ./bin/ssi-scheduler
SSI Scheduler starting...
Scheduling PodGroup `gang` of size 3 with priority 'high'...
Acquired 3 vacant node(s) for PodGroup gang

*** PodGroup Status Summary ***
PodGroup ID: gang
PodGroup Size: 3
PodGroup Priority: high
Pod Details:
  Pod gangster-1 (in namespace default) is scheduled on node kind-worker
  Pod gangster-2 (in namespace default) is scheduled on node kind-worker2
  Pod gangster-3 (in namespace default) is scheduled on node kind-worker3


Scheduling PodGroup `pod-high` of size 1 with priority 'high'...
Acquired 1 vacant node(s) for PodGroup pod-high

*** PodGroup Status Summary ***
PodGroup ID: pod-high
PodGroup Size: 1
PodGroup Priority: high
Pod Details:
  Pod pod-high (in namespace default) is scheduled on node kind-worker4


Scheduling PodGroup `team` of size 4 with priority 'medium'...
Insufficient vacant nodes available for PodGroup team
Scheduling PodGroup `team` of size 4 with priority 'medium'...
Insufficient vacant nodes available for PodGroup team

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
crewmate-1      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
crewmate-2      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
crewmate-3      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
gangster-1      1/1     Running   0          20s   10.244.2.2   kind-worker    <none>           <none>
gangster-2      1/1     Running   0          20s   10.244.1.2   kind-worker2   <none>           <none>
gangster-3      1/1     Running   0          20s   10.244.3.2   kind-worker3   <none>           <none>
pod-high        1/1     Running   0          20s   10.244.4.2   kind-worker4   <none>           <none>
pod-none        0/1     Pending   0          20s   <none>       <none>         <none>           <none>
pod-undefined   0/1     Pending   0          20s   <none>       <none>         <none>           <none>
teammate-1      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
teammate-2      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
teammate-3      0/1     Pending   0          20s   <none>       <none>         <none>           <none>
teammate-4      0/1     Pending   0          20s   <none>       <none>         <none>           <none>

=== Remove pod ===
pod "gangster-1" deleted from default namespace
pod "gangster-2" deleted from default namespace
pod "gangster-3" deleted from default namespace
Scheduling PodGroup `team` of size 4 with priority 'medium'...
Insufficient vacant nodes available for PodGroup team
Scheduling PodGroup `team` of size 4 with priority 'medium'...
Insufficient vacant nodes available for PodGroup team
Scheduling PodGroup `team` of size 4 with priority 'medium'...
Insufficient vacant nodes available for PodGroup team

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
crewmate-1      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
crewmate-2      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
crewmate-3      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
pod-high        1/1     Running   0          32s   10.244.4.2   kind-worker4   <none>           <none>
pod-none        0/1     Pending   0          32s   <none>       <none>         <none>           <none>
pod-undefined   0/1     Pending   0          32s   <none>       <none>         <none>           <none>
teammate-1      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
teammate-2      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
teammate-3      0/1     Pending   0          32s   <none>       <none>         <none>           <none>
teammate-4      0/1     Pending   0          32s   <none>       <none>         <none>           <none>

=== Remove pod ===
pod "pod-high" deleted from default namespace
Scheduling PodGroup `team` of size 4 with priority 'medium'...
Acquired 4 vacant node(s) for PodGroup team

*** PodGroup Status Summary ***
PodGroup ID: team
PodGroup Size: 4
PodGroup Priority: medium
Pod Details:
  Pod teammate-1 (in namespace default) is scheduled on node kind-worker
  Pod teammate-2 (in namespace default) is scheduled on node kind-worker2
  Pod teammate-3 (in namespace default) is scheduled on node kind-worker3
  Pod teammate-4 (in namespace default) is scheduled on node kind-worker4


Scheduling PodGroup `crew` of size 3 with priority 'low'...
Insufficient vacant nodes available for PodGroup crew

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
crewmate-1      0/1     Pending   0          43s   <none>       <none>         <none>           <none>
crewmate-2      0/1     Pending   0          43s   <none>       <none>         <none>           <none>
crewmate-3      0/1     Pending   0          43s   <none>       <none>         <none>           <none>
pod-none        0/1     Pending   0          43s   <none>       <none>         <none>           <none>
pod-undefined   0/1     Pending   0          43s   <none>       <none>         <none>           <none>
teammate-1      1/1     Running   0          43s   10.244.2.3   kind-worker    <none>           <none>
teammate-2      1/1     Running   0          43s   10.244.1.3   kind-worker2   <none>           <none>
teammate-3      1/1     Running   0          43s   10.244.3.3   kind-worker3   <none>           <none>
teammate-4      1/1     Running   0          43s   10.244.4.3   kind-worker4   <none>           <none>

=== Remove pod ===
pod "teammate-1" deleted from default namespace
pod "teammate-2" deleted from default namespace
pod "teammate-3" deleted from default namespace
pod "teammate-4" deleted from default namespace
Scheduling PodGroup `crew` of size 3 with priority 'low'...
Acquired 3 vacant node(s) for PodGroup crew

*** PodGroup Status Summary ***
PodGroup ID: crew
PodGroup Size: 3
PodGroup Priority: low
Pod Details:
  Pod crewmate-1 (in namespace default) is scheduled on node kind-worker
  Pod crewmate-2 (in namespace default) is scheduled on node kind-worker2
  Pod crewmate-3 (in namespace default) is scheduled on node kind-worker3


Scheduling PodGroup `pod-none` of size 1 with priority 'none'...
Acquired 1 vacant node(s) for PodGroup pod-none

*** PodGroup Status Summary ***
PodGroup ID: pod-none
PodGroup Size: 1
PodGroup Priority: none
Pod Details:
  Pod pod-none (in namespace default) is scheduled on node kind-worker4


Scheduling PodGroup `pod-undefined` of size 1 with priority 'none'...
Insufficient vacant nodes available for PodGroup pod-undefined

=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
crewmate-1      1/1     Running   0          54s   10.244.2.4   kind-worker    <none>           <none>
crewmate-2      1/1     Running   0          54s   10.244.1.4   kind-worker2   <none>           <none>
crewmate-3      1/1     Running   0          54s   10.244.3.4   kind-worker3   <none>           <none>
pod-none        1/1     Running   0          54s   10.244.4.4   kind-worker4   <none>           <none>
pod-undefined   0/1     Pending   0          54s   <none>       <none>         <none>           <none>

=== Remove pod ===
pod "pod-none" deleted from default namespace
Scheduling PodGroup `pod-undefined` of size 1 with priority 'none'...
Insufficient vacant nodes available for PodGroup pod-undefined
Scheduling PodGroup `pod-undefined` of size 1 with priority 'none'...
Acquired 1 vacant node(s) for PodGroup pod-undefined

*** PodGroup Status Summary ***
PodGroup ID: pod-undefined
PodGroup Size: 1
PodGroup Priority: none
Pod Details:
  Pod pod-undefined (in namespace default) is scheduled on node kind-worker4



=== Pod Status ===
NAME            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
crewmate-1      1/1     Running   0          65s   10.244.2.4   kind-worker    <none>           <none>
crewmate-2      1/1     Running   0          65s   10.244.1.4   kind-worker2   <none>           <none>
crewmate-3      1/1     Running   0          65s   10.244.3.4   kind-worker3   <none>           <none>
pod-undefined   1/1     Running   0          65s   10.244.4.5   kind-worker4   <none>           <none>

=== Scheduling Events (verify priority order) ===
58s         Normal   Scheduled   pod/gangster-1      Successfully scheduled pod gangster-1 (namespace default) on node kind-worker
57s         Normal   Scheduled   pod/gangster-2      Successfully scheduled pod gangster-2 (namespace default) on node kind-worker2
56s         Normal   Scheduled   pod/gangster-3      Successfully scheduled pod gangster-3 (namespace default) on node kind-worker3
55s         Normal   Scheduled   pod/pod-high        Successfully scheduled pod pod-high (namespace default) on node kind-worker4
29s         Normal   Scheduled   pod/teammate-1      Successfully scheduled pod teammate-1 (namespace default) on node kind-worker
28s         Normal   Scheduled   pod/teammate-2      Successfully scheduled pod teammate-2 (namespace default) on node kind-worker2
27s         Normal   Scheduled   pod/teammate-3      Successfully scheduled pod teammate-3 (namespace default) on node kind-worker3
26s         Normal   Scheduled   pod/teammate-4      Successfully scheduled pod teammate-4 (namespace default) on node kind-worker4
20s         Normal   Scheduled   pod/crewmate-1      Successfully scheduled pod crewmate-1 (namespace default) on node kind-worker
19s         Normal   Scheduled   pod/crewmate-2      Successfully scheduled pod crewmate-2 (namespace default) on node kind-worker2
18s         Normal   Scheduled   pod/crewmate-3      Successfully scheduled pod crewmate-3 (namespace default) on node kind-worker3
17s         Normal   Scheduled   pod/pod-none        Successfully scheduled pod pod-none (namespace default) on node kind-worker4
6s          Normal   Scheduled   pod/pod-undefined   Successfully scheduled pod pod-undefined (namespace default) on node kind-worker4
Cleaning up...
kind delete cluster
rm -rf bin/
```

## Additional improvements

- ~~Gang-scheduling: Pods with the same pod-group annotation should all get scheduled (or all NOT if insufficient nodes)~~ *Update: Implemented!
- Retry mechanism: Current retry is inherent to the unending for-loop; a more sophisticated retry algorithm could entail "booting" already-scheduled/already-bound, lower priority pods when higher priority pods are newly launched but no nodes are available
- Performance: Current bottleneck is definitely O(nlogn) sorting of pods by priority; using a combination of Informer (K8s construct for event-driven pod handling) + priority queue (eliminates redundant re-sorts) could enhance scheduling speed, especially as pod count and cluster size grow
