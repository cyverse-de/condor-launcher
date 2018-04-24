FROM golang:1.7-alpine

RUN apk update && apk add git

RUN go get github.com/jstemmer/go-junit-report

COPY . /go/src/github.com/cyverse-de/logcabin

CMD go test -v github.com/cyverse-de/logcabin | tee /dev/stderr | go-junit-report
