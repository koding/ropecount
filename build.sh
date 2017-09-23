#!/bin/bash

set -euo pipefail

export GOPATH=$(
	cd "$(dirname "$0")"/../../../.. #src/github.com/koding/ropecount
	pwd
)
export GOBIN=${GOBIN:-${GOPATH}/bin}
export GO_LDFLAGS=""
export GO_TAGS=""

go-install() {
	go install -v -tags "${GO_TAGS}" -ldflags "${GO_LDFLAGS}" $*
}

export COMMANDS=(
	$(go list github.com/koding/ropecount/... | grep -v vendor)
)

go-install ${COMMANDS[@]}
