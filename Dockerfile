FROM golang:1.17.3-alpine as build-env
RUN GO111MODULE=on go get -v github.com/projectdiscovery/simplehttpserver/cmd/simplehttpserver

FROM alpine:latest
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=build-env /go/bin/simplehttpserver /usr/local/bin/simplehttpserver
ENTRYPOINT ["simplehttpserver"]