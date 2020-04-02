#!/usr/bin/env bash

source .plsdo.sh

banner="Consul - project tasks"

help[check]="Run 'shellcheck' on all files."
check() {
    if ! command -v shellcheck > /dev/null; then
        _plsdo_error "Missing shellcheck."
        _plsdo_error "See https://github.com/koalaman/shellcheck#installing"
        return 3
    fi
    # TODO: add other files
    shellcheck --severity=style --external-sources ./do
}

help[go-mod-tidy]="Run 'go mod tidy' on all go modules."
go-mod-tidy() {
    echo "--> Running go mod tidy"
    go mod tidy
    (cd sdk && go mod tidy)
    (cd api && go mod tidy)
}

help[update-vendor]="Update ./vendor after making changing dependencies."
update-vendor() {
    go-mod-tidy
    echo "--> Running go mod vendor"
    go mod vendor
    echo "--> Removing vendoring of our own nested modules"
    rm -rf vendor/github.com/hashicorp/consul
    grep -v "hashicorp/consul/" < vendor/modules.txt > vendor/modules.txt.new
    mv vendor/modules.txt.new vendor/modules.txt
}

help[lint]="Run 'golangci-lint' on all files.

Environment Variables:

  GOTAGS - used as Go build tags
"
lint() {
    ${GOTAGS=}
    echo "--> Running go golangci-lint"
    golangci-lint run --build-tags "${GOTAGS}"
    (cd api && golangci-lint run --build-tags "${GOTAGS}")
    (cd sdk && golangci-lint run --build-tags "${GOTAGS}")
}

help[binary]="Build the consul binary.

Defaults to building a binary for the local GOOS and GOARCH.

TODO: env vars
"
binary() {
    local target=bin/consul
    if [ -n "${GOOS-}" ]; then
        target="bin/consul-${GOOS}-${GOARCH}"
    fi
    go build -o "$target" -ldflags "$(_go_build_ldflags)" .
}

_go_build_ldflags() {
    local commit; commit=$(git rev-parse --short HEAD)
    local dirty;   [ -z "$(git status --porcelain)" ] || dirty="+CHANGES"
    local desc;   desc="$(git describe --tags --always --match 'v*')"
    local import=github.com/hashicorp/consul/version
    echo "-X ${import}.GitCommit=${commit}${dirty} -X ${import}.GitDescribe=${desc}"
}

binary-all() {
    # TODO: use parallel
    GOOS=linux  GOARCH=amd64 binary
    GOOS=darwin GOARCH=amd64 binary
}

cross() { binary-all; }

_plsdo_run "$@"
