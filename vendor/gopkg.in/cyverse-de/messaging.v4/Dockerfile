FROM golang:1.7-alpine

RUN apk update && apk add git

RUN go get github.com/jstemmer/go-junit-report

RUN go get github.com/olebedev/config
RUN go get github.com/cyverse-de/logcabin
RUN go get github.com/cyverse-de/model
RUN go get github.com/cyverse-de/configurate
RUN go get github.com/streadway/amqp

COPY . /go/src/github.com/cyverse-de/messaging

CMD go test -v github.com/cyverse-de/messaging | tee /dev/stderr | go-junit-report
