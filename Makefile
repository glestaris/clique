.PHONY: all test clean

all: ./clique-agent

test:
	ginkgo -r -p -race

./clique-agent:
	go build ./cmd/clique-agent/...

clean:
	rm -f ./clique-agent
