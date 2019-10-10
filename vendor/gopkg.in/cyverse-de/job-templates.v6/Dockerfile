FROM golang:1.9-alpine

RUN apk update && apk add git

RUN go get github.com/jstemmer/go-junit-report
RUN go get github.com/golang/dep/cmd/dep

COPY . /go/src/github.com/cyverse-de/job-templates

RUN cd /go/src/github.com/cyverse-de/job-templates && dep ensure

CMD go test -v github.com/cyverse-de/job-templates | tee /dev/stderr | go-junit-report
