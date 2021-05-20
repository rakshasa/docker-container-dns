FROM golang:1.16.4-alpine3.13 AS builder

WORKDIR /build

RUN set -xe; \
  apk add --no-cache \
    curl \
    gcc \
    musl-dev

COPY . ./

ENV GO111MODULE=on

RUN set -xe; \
  go build \
  -v \
  -mod=readonly \
  -mod=vendor \
  -ldflags "-linkmode external -extldflags '-static -fno-PIC' -s -w"


FROM scratch

COPY --from=builder /build/docker-container-dns /

ENTRYPOINT ["/docker-container-dns"]
