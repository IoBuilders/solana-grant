# syntax=docker/dockerfile:1

# Builds the naryo-go microservice binary. Build context MUST be the repo root, not server/ —
# this is a go.work workspace where server/go.mod `replace`s core, persistence-gorm, and
# broadcaster-http with local sibling paths (../core, etc.), so the build needs every module
# listed in go.work on disk together (go.work resolution fails if any `use`d module directory is
# missing, even one server doesn't import yet — e.g. broadcaster-kafka):
#
#   docker build -f server/Dockerfile -t naryo-server .

FROM golang:1.26.1-alpine AS builder

WORKDIR /workspace

COPY go.work go.work.sum ./
COPY core/ core/
COPY broadcaster-http/ broadcaster-http/
COPY broadcaster-kafka/ broadcaster-kafka/
COPY broadcaster-rabbitmq/ broadcaster-rabbitmq/
COPY persistence-gorm/ persistence-gorm/
COPY server/ server/

RUN go build -o /out/naryo-server ./server/cmd/server

FROM alpine:3.21

# curl+bash are needed only to install the atlas CLI (below); atlasexec (used by
# persistencegormsetup.Setup to apply DB migrations at startup) shells out to it.
RUN apk add --no-cache ca-certificates curl bash && \
    curl -sSf https://atlasgo.sh | sh && \
    adduser -D -u 10001 naryo
USER naryo

COPY --from=builder /out/naryo-server /usr/local/bin/naryo-server
# Baked in only so CONFIG_PATH (e.g. `make example-quickstart`) can point at an example config
# by its repo-relative path; the container's default WORKDIR (/) keeps that path valid.
COPY --chown=naryo:naryo examples/ examples/

EXPOSE 8060

ENTRYPOINT ["naryo-server"]
