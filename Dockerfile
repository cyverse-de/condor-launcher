FROM golang:1.7

ARG git_commit=unknown
LABEL org.cyverse.git-ref="$git_commit"

COPY . /go/src/github.com/cyverse-de/condor-launcher
RUN go install github.com/cyverse-de/condor-launcher

ENTRYPOINT ["condor-launcher"]
CMD ["--help"]
