go get k8s.io/code-generator || true
cd ${GOPATH%%:*}/src/k8s.io/code-generator
go install ./cmd/...