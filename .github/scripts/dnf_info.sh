#!/bin/bash

# this is meant to be run inside a container that supports the dnf package manager (eg. Fedora)
function main {
  local pkg=${1:-}
  local version=${2:-}

  if [[ -z "${version}" ]]; then
    echo "ERROR: version argument is required"
    exit 1
  fi

  dnf install -y dnf-plugins-core
  dnf config-manager --add-repo https://rpm.releases.hashicorp.com/fedora/hashicorp.repo

  # should show latest OSS and Ent versions
  dnf info "${pkg}"

  dnf --showduplicates list "${pkg}" | grep "${version}"
}

main "$@"
