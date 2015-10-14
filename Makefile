.PHONY: all clean

all: ./clique-agent

./clique-agent:
	go build ./cmd/clique-agent/...

clean:
	rm -f ./clique-agent 
