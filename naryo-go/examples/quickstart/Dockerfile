# syntax=docker/dockerfile:1

# Builds the quickstart example's receiver: a standalone Go file (stdlib only, no go.mod of its
# own — it isn't part of the go.work workspace). Build context must be the repo root:
#
#   docker build -f examples/quickstart/Dockerfile -t quickstart-receiver .

FROM golang:1.26.1-alpine AS builder

WORKDIR /workspace
COPY examples/quickstart/http_server.go .
RUN go build -o /out/http-server http_server.go

FROM alpine:3.21

RUN adduser -D -u 10001 receiver
USER receiver

COPY --from=builder /out/http-server /usr/local/bin/http-server

EXPOSE 8081

ENTRYPOINT ["http-server"]
