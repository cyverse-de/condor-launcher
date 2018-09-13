FROM golang:1.10.2

COPY . /go/src/github.com/cyverse-de/messaging

RUN curl -LO https://raw.githubusercontent.com/golang/dep/master/install.sh && \
	chmod +x install.sh && \
	./install.sh

WORKDIR /go/src/github.com/cyverse-de/messaging

RUN go get github.com/jstemmer/go-junit-report

RUN dep ensure

CMD go test -v github.com/cyverse-de/messaging | tee /dev/stderr | go-junit-report
