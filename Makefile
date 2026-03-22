KUBECONFIG ?= $(HOME)/.kube/config

setup-go:
	test -f go.mod || go mod init ssi-scheduler
	go get k8s.io/client-go@latest
	go get k8s.io/api@latest
	go get k8s.io/apimachinery@latest
	go mod tidy

start-cluster:
	kind create cluster --config kind-config.yaml

build:
	go build -o bin/ssi-scheduler .

run: build
	KUBECONFIG=$(KUBECONFIG) ./bin/ssi-scheduler

stop-cluster:
	kind delete cluster

clean:
	rm -rf bin/
