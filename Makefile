.PHONY: all test clean

all: ./clique-agent

test:
	ginkgo -p acceptance
	ginkgo -r -p -race -skipPackage acceptance

./clique-agent:
	go build ./cmd/clique-agent/...

clean:
	rm -f ./clique-agent
