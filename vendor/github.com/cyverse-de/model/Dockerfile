FROM golang:1.7-alpine

RUN apk update && apk add git

RUN go get github.com/jstemmer/go-junit-report

RUN go get github.com/spf13/viper
RUN go get github.com/cyverse-de/configurate

COPY . /go/src/github.com/cyverse-de/model

CMD go test -v github.com/cyverse-de/model | tee /dev/stderr | go-junit-report
