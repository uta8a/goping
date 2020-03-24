# for build
FROM golang:alpine3.11 AS builder

# git for Go Modules
RUN apk update \
  && apk add git

RUN mkdir /app
WORKDIR /app
COPY main.go .
COPY go.mod .
COPY go.sum .
COPY handler/ ./handler

RUN go mod download

RUN go build -o /api

# production
FROM alpine:3.11

COPY --from=builder /api .
ENTRYPOINT ["/api"]
