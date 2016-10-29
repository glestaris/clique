.PHONY: linux clean \
	all go-vet test \
	deps update-deps \
	linux

all: ./build/clique-agent

go-vet:
	go vet `go list ./... | grep -v vendor`

test:
	go install github.com/onsi/ginkgo/ginkgo
	ginkgo -randomizeAllSpecs -p acceptance
	ginkgo -randomizeAllSpecs -randomizeSuites -r -p -race -skipPackage acceptance,ctl,vendor
	ginkgo -randomizeAllSpecs ctl

deps:
	glide install

update-deps:
	glide update

./build/clique-agent:
	go build -o ./build/clique-agent ./cmd/clique-agent/...

./build/linux/clique-agent:
	mkdir -p build/linux
	GOOS=linux go build -o ./build/linux/clique-agent ./cmd/clique-agent/...

linux: ./build/linux/clique-agent

clean:
	rm -Rf ./build
