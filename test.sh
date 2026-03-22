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
  name: pod-low
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
  name: pod-medium
  labels:
    priority: medium
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
    | grep -E "pod-high|pod-medium|pod-low|pod-none|pod-undefined"
