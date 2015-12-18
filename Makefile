.PHONY: all setup test ice-clique-docker linux linux-test clean

all: ./build/clique-agent

setup:
	go get github.com/tools/godep
	godep restore

test: setup
	go install github.com/onsi/ginkgo/ginkgo
	ginkgo -p acceptance
	ginkgo -r -p -race -skipPackage acceptance

./build/clique-agent: setup
	go build -o ./build/clique-agent ./cmd/clique-agent/...

ice-clique-docker:
	docker build -t glestaris/ice-clique .

./build/linux/clique-agent: ice-clique-docker
	docker run --name="ice-clique-builder" \
		glestaris/ice-clique \
		make
	mkdir -p build/linux
	docker cp \
		ice-clique-builder:/go/src/github.com/glestaris/ice-clique/build/clique-agent \
		./build/linux/clique-agent
	docker rm ice-clique-builder

linux: ./build/linux/clique-agent

linux-test: ice-clique-docker
	docker run --name="ice-clique-tester" \
		glestaris/ice-clique \
		make test
	docker rm ice-clique-tester

clean:
	rm -Rf ./build
	docker rmi glestaris/ice-clique
