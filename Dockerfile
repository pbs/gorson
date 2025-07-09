FROM golang:1.24.4-alpine

# borrowed with gratitude from confd
# https://github.com/kelseyhightower/confd/blob/master/Dockerfile.build.alpine?at=09f6676

WORKDIR /app
RUN apk add --no-cache bash gcc git musl-dev

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download
