# Pinning apk package versions breaks the build once Alpine drops them; the base image tag is the pin.
# hadolint global ignore=DL3018
FROM golang:1.27.3-alpine AS build-env
RUN apk add --no-cache git
WORKDIR /app
COPY . /app
RUN go mod download && go build ./cmd/simplehttpserver

FROM alpine:3.24
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=build-env /app/simplehttpserver /usr/local/bin/simplehttpserver
ENTRYPOINT ["simplehttpserver"]
