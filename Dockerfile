FROM golang:1.8

WORKDIR /go/src/github.com/koding/ropecount

ADD . .

RUN /go/src/github.com/koding/ropecount/build.sh