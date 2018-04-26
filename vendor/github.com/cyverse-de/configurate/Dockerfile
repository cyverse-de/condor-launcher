FROM golang:1.6-alpine

RUN apk update && apk add git

RUN go get github.com/spf13/viper

COPY . /go/src/github.com/cyverse-de/configurate

CMD ["go", "test", "github.com/cyverse-de/configurate"]
