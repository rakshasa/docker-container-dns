#!/bin/bash

PROJECT_PATH="$(cd "$(cd "$( dirname "${BASH_SOURCE[0]}" )" && git rev-parse --show-toplevel)" >/dev/null 2>&1 && pwd)"

MODULE_PATH="github.com/rakshasa/docker-container-dns"

GO_MODULES=(
  "github.com/docker/docker@v20.10.6"
)

set -xeu

cd "${PROJECT_PATH}"

rm -f ./go.{mod,sum}
go clean -cache

go mod init "${MODULE_PATH}"

for mod in "${GO_MODULES[@]}"; do
  go get -u -v "${mod}"
done

go mod tidy -v
go mod vendor -v
