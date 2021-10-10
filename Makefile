
export GO111MODULE=on

.PHONY: test
test:
	go test ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/kubectl-capacity github.com/akrzos/kubeSize/

.PHONY: fmt
fmt:
	go fmt ./cmd/...

.PHONY: vet
vet:
	go vet ./cmd/...

.PHONY: kubernetes-deps
kubernetes-deps:
	go get k8s.io/client-go@v0.21.2
	go get k8s.io/api@kubernetes-1.21.0
	go get k8s.io/apimachinery@kubernetes-1.21.0
	go get k8s.io/cli-runtime@kubernetes-1.21.0

.PHONY: setup
setup:
	make -C setup