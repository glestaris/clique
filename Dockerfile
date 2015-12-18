FROM golang:1.5

ADD . /go/src/github.com/glestaris/ice-clique
WORKDIR /go/src/github.com/glestaris/ice-clique
RUN go get github.com/tools/godep && godep restore

CMD ["/bin/bash"]
