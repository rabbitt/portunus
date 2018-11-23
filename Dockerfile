FROM golang:1.11 as build

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN curl -s -L -o /tmp/goreleaser.tgz \
    "https://github.com/goreleaser/goreleaser/releases/download/v0.46.3/goreleaser_$(uname -s)_$(uname -m).tar.gz" \
    && tar -xf /tmp/goreleaser.tgz -C /usr/local/bin

WORKDIR /go/src/github.com/rabbitt/portunus
COPY . /go/src/github.com/rabbitt/portunus
RUN find /go && make clean && make

FROM scratch
COPY --from=build /go/src/github.com/rabbitt/portunus/bin/portunus /portunus

ENTRYPOINT ["/portunus"]
CMD [ "--help" ]
