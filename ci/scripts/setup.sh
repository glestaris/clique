#!/bin/bash
set -e -x

export GOPATH="/go"
export PATH="/go/bin:$PATH"

go get github.com/tools/godep
cd ice-clique/
godep restore

go install github.com/onsi/ginkgo/ginkgo
