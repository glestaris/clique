.PHONY: linux clean setup \
	all test save-deps \
	linux linux-test linux-save-deps \
  clique-docker

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

clique-docker:
	docker build -t glestaris/clique .

./build/linux/clique-agent: clique-docker
	docker run --name="clique-builder" \
		glestaris/clique \
		make
	mkdir -p build/linux
	docker cp \
		clique-builder:/go/src/github.com/glestaris/clique/build/clique-agent \
		./build/linux/clique-agent
	docker rm clique-builder

linux: ./build/linux/clique-agent

linux-test: clique-docker
	docker run --name="clique-tester" \
		glestaris/clique \
		make test
	docker rm clique-tester

linux-save-deps: clique-docker
	docker run --name="clique-deps-saver" \
		glestaris/clique \
		make save-deps
	docker cp \
		clique-deps-saver:/go/src/github.com/glestaris/clique/Godeps \
		.
	docker rm clique-deps-saver

clean:
	rm -Rf ./build
