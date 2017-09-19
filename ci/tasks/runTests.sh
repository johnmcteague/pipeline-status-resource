#!/bin/bash -e

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:$PATH

go get github.com/Masterminds/glide
WORKING_DIR=$GOPATH/src/github.com/pivotalservices/pipeline-status-resource
mkdir -p ${WORKING_DIR}
cp -R source/* ${WORKING_DIR}/.
cd ${WORKING_DIR}

go get github.com/onsi/ginkgo

glide install
#go test $(glide nv) -v
CGO_ENABLED=1 ginkgo -race -r -p "$@"
