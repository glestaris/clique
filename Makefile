.PHONY: linux clean setup \
	all test save-deps \
	linux linux-test linux-save-deps \
  ice-clique-docker

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

linux-save-deps: ice-clique-docker
	docker run --name="ice-clique-deps-saver" \
		glestaris/ice-clique \
		make save-deps
	docker cp \
		ice-clique-deps-saver:/go/src/github.com/glestaris/ice-clique/Godeps \
		.
	docker rm ice-clique-deps-saver

clean:
	rm -Rf ./build
