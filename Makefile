.PHONY: linux clean setup \
	all test save-deps \
	linux

all: ./build/clique-agent

setup:
	go get github.com/tools/godep
	godep restore

test: setup
	go install github.com/onsi/ginkgo/ginkgo
	ginkgo -randomizeAllSpecs -p acceptance
	ginkgo -randomizeAllSpecs -randomizeSuites -r -p -race -skipPackage acceptance,ctl
	ginkgo -randomizeAllSpecs ctl

save-deps:
	go get github.com/tools/godep
	godep save ./...

./build/clique-agent: setup
	go build -o ./build/clique-agent ./cmd/clique-agent/...

./build/linux/clique-agent: setup
	mkdir -p build/linux
	GOOS=linux go build -o ./build/linux/clique-agent ./cmd/clique-agent/...

linux: ./build/linux/clique-agent

clean:
	rm -Rf ./build
