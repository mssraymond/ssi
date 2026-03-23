#!/bin/bash
set -e

clear && clear

echo "=== Go + Cluster Setup ==="

make setup-go start-cluster
sleep 3
make run 2>/dev/null &
SCHEDULER_PID=$!
trap "echo 'Cleaning up...'; \
      kill $SCHEDULER_PID 2>/dev/null || true; \
      wait $SCHEDULER_PID 2>/dev/null || true; \
      make -k stop-cluster clean 2>/dev/null || true" EXIT

echo ""
echo "=== SSI Scheduler Test ==="

# Launch 5 pods, schedule priority-labeled ones according to priority order, followed by those without priority labels (1 unscheduled)
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: pod-undefined
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-none
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: crewmate-1
  annotations:
    ssi.scheduler/pod-group: crew
  labels:
    priority: low
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: crewmate-2
  annotations:
    ssi.scheduler/pod-group: crew
  labels:
    priority: low
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: crewmate-3
  annotations:
    ssi.scheduler/pod-group: crew
  labels:
    priority: low
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: pod-high
  labels:
    priority: high
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: teammate-1
  annotations:
    ssi.scheduler/pod-group: team
  labels:
    priority: medium
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: teammate-2
  annotations:
    ssi.scheduler/pod-group: team
  labels:
    priority: medium
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: teammate-3
  annotations:
    ssi.scheduler/pod-group: team
  labels:
    priority: medium
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: teammate-4
  annotations:
    ssi.scheduler/pod-group: team
  labels:
    priority: medium
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: gangster-1
  annotations:
    ssi.scheduler/pod-group: gang
  labels:
    priority: low
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: gangster-2
  annotations:
    ssi.scheduler/pod-group: gang
  labels:
    priority: medium
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
---
apiVersion: v1
kind: Pod
metadata:
  name: gangster-3
  annotations:
    ssi.scheduler/pod-group: gang
  labels:
    priority: high
spec:
  schedulerName: ssi-scheduler
  containers:
    - name: nginx
      image: nginx
EOF

sleep 20
echo ""
echo "=== Pod Status ==="
kubectl get pods -o wide

# Remove pods from 'gang' PodGroup
echo ""
echo "=== Remove pod ==="
kubectl delete pod gangster-1 gangster-2 gangster-3

sleep 10
echo ""
echo "=== Pod Status ==="
kubectl get pods -o wide

# Remove pod 'pod-high' to make room for PodGroup 'team'
echo ""
echo "=== Remove pod ==="
kubectl delete pod pod-high

sleep 10
echo ""
echo "=== Pod Status ==="
kubectl get pods -o wide

# Remove pods from 'team' PodGroup
echo ""
echo "=== Remove pod ==="
kubectl delete pod teammate-1 teammate-2 teammate-3 teammate-4

sleep 10
echo ""
echo "=== Pod Status ==="
kubectl get pods -o wide

# Remove the scheduled "priority-less" pod, giving the straggler pod a chance to be scheduled
echo ""
echo "=== Remove pod ==="
kubectl delete pod pod-none

sleep 10
echo ""
echo "=== Pod Status ==="
kubectl get pods -o wide

echo ""
echo "=== Scheduling Events (verify priority order) ==="
kubectl get events \
    --field-selector reason=Scheduled \
    --sort-by='.lastTimestamp' \
    | grep -E "pod-high|pod-medium|pod-low|pod-none|pod-undefined|gangster-1|gangster-2|gangster-3|teammate-1|teammate-2|teammate-3|teammate-4|crewmate-1|crewmate-2|crewmate-3"
