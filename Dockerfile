FROM golang:1.9.1

WORKDIR /go/src/github.com/koding/ropecount

ADD . .

RUN /go/src/github.com/koding/ropecount/build.sh