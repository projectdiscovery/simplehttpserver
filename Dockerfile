FROM golang:1.18-alpine as build-env
RUN go install -v github.com/projectdiscovery/simplehttpserver/cmd/simplehttpserver@latest

FROM alpine:latest
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=build-env /go/bin/simplehttpserver /usr/local/bin/simplehttpserver
ENTRYPOINT ["simplehttpserver"]