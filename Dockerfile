FROM golang:1.5

ADD . /go/src/github.com/glestaris/clique
WORKDIR /go/src/github.com/glestaris/clique
RUN go get github.com/tools/godep && godep restore

CMD ["/bin/bash"]
