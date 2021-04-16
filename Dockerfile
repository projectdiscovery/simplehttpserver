FROM golang:1.16-alpine AS builder
RUN apk add --no-cache git
RUN GO111MODULE=on go get -v github.com/projectdiscovery/simplehttpserver/cmd/simplehttpserver

FROM alpine:latest
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=builder /go/bin/simplehttpserver /usr/local/bin/

ENTRYPOINT ["simplehttpserver"]
