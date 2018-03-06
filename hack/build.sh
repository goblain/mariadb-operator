#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o mariadb-operator -installsuffix cgo ./cmd/


## function go_build {
## 	if [ ! -z ${GOINSTALL+x} ] && [ "${GOINSTALL}" = "y" ]
## 	then
##   		GOBIN=${bin_dir} GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go install -installsuffix cgo -ldflags "$go_ldflags" ./cmd/${1}/
##   		mv ${bin_dir}/${1} ${bin_dir}/etcd-${1}
## 	else
##   		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${bin_dir}/etcd-${1} -installsuffix cgo -ldflags "$go_ldflags" ./cmd/${1}/
## 	fi
## }
## 
## if ! which go > /dev/null; then
## 	echo "golang needs to be installed"
## 	exit 1
## fi
## 
## if ! which docker > /dev/null; then
## 	echo "docker needs to be installed"
## 	exit 1
## fi
## 
## : ${IMAGE:?"Need to set IMAGE, e.g. gcr.io/coreos-k8s-scale-testing/etcd-operator"}
## 
## GIT_SHA=`git rev-parse --short HEAD || echo "GitNotFound"`
## 
## bin_dir="$(pwd)/_output/bin"
## mkdir -p ${bin_dir} || true


## ldKVPairs="github.com/goblain/mariadb-operator/pkg/util/k8sutil.BackupImage=${IMAGE}"
## gitHash="github.com/goblain/mariadb-operator/version.GitSHA=${GIT_SHA}"

## go_ldflags="-X ${ldKVPairs} -X ${gitHash}"

## go_build operator
## go_build backup

## docker build --tag "${IMAGE}" -f hack/build/operator/Dockerfile . 1>/dev/null
## # For gcr users, do "gcloud docker -a" to have access.
## docker push "${IMAGE}" 1>/dev/null

