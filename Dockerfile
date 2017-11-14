FROM golang:1.9.1

WORKDIR /go/src/github.com/ropelive/count

ADD . .

RUN /go/src/github.com/ropelive/count/build.sh