FROM golang:1.7

ADD . /go/src/github.com/glestaris/clique
WORKDIR /go/src/github.com/glestaris/clique
RUN go get github.com/tools/godep && godep restore

CMD ["/bin/bash"]
