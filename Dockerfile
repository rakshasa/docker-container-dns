FROM golang:1.16.4-alpine3.13 AS builder

WORKDIR /build

RUN set -xe; \
  apk add --no-cache \
    curl \
    gcc \
    musl-dev

COPY . ./

ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV GOOS=linux

RUN set -xe; \
  go build \
  -v \
  -mod=readonly \
  -mod=vendor \
  -ldflags "-linkmode external -extldflags '-static -fno-PIC' -s -w"


FROM scratch

COPY --from=builder /build/docker-container-dns /

ENTRYPOINT ["/docker-container-dns"]
