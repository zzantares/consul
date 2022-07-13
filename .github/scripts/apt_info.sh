#!/bin/bash

# this is meant to be run inside a container that supports the apt package manager (eg. debian, ubuntu)
function main {
  local pkg=${1:-}
  local version=${2:-}
  local platform

  if [[ -z "${pkg}" ]]; then
    echo "ERROR: pkg argument is required"
    exit 1
  fi

  if [[ -z "${version}" ]]; then
    echo "ERROR: version argument is required"
    exit 1
  fi

  apt update && apt install -y software-properties-common curl
  curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add -

  platform=$(uname -m)
  case "${platform}" in
    x86_64)
      apt-add-repository -y "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
      ;;
    aarch64)
      apt-add-repository -y "deb [arch=arm64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
      ;;
    *)
      echo "Unknown platform: ${platform}, must be either x86_64 or aarch64 (arm64)"
      exit 1
      ;;
  esac

  apt-get update && apt show "${pkg}"
  apt list -a "${pkg}" | grep ${version} 
}

main "$@"
