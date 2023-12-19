FROM golang:1.21-alpine

RUN apk add --no-cache git wget

COPY . /go/src/github.com/cyverse-de/condor-launcher
WORKDIR /go/src/github.com/cyverse-de/condor-launcher
RUN go install github.com/cyverse-de/condor-launcher

ENTRYPOINT ["condor-launcher"]
CMD ["--help"]

ARG git_commit=unknown
ARG version="2.20.0"
ARG descriptive_version=unknown

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
LABEL org.cyverse.descriptive-version="$descriptive_version"
LABEL org.label-schema.vcs-ref="$git_commit"
LABEL org.label-schema.vcs-url="https://github.com/cyverse-de/condor-launcher"
LABEL org.label-schema.version="$descriptive_version"
