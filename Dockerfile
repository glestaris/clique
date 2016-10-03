FROM golang:1.7

ADD . /go/src/github.com/ice-stuff/clique
WORKDIR /go/src/github.com/ice-stuff/clique
RUN go get github.com/tools/godep && godep restore

CMD ["/bin/bash"]
