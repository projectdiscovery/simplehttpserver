FROM golang:1.16-alpine AS builder
RUN apk add --no-cache git
RUN GO111MODULE=auto go get -u -v github.com/projectdiscovery/simplehttpserver/cmd/simplehttpserver

FROM alpine:latest
COPY --from=builder /go/bin/simplehttpserver /usr/local/bin/

ENTRYPOINT ["simplehttpserver"]
