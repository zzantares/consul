#!/bin/bash

set -euo pipefail

# verify_linux_packages.sh verifies that a given version of either Consul or Consul Enterprise
# is present in the Linux package repository for a given distro of Linux. It does this by invoking
# an associated Docker image and executing package manager commands to query for the correct version.

function usage {
    echo "usage:"
    echo "  verify_linux_packages.sh <product> <flavor> <version>"
    echo
    echo "example:"
    echo "  verify_linux_packages.sh consul-enterprise ubuntu 1.12.2"

}

PRODUCT_OSS="consul"
PRODUCT_ENT="consul-enterprise"
IMAGE_UBUNTU="ubuntu:22.04"
IMAGE_DEBIAN="debian:bullseye"
IMAGE_CENTOS="centos:7"
IMAGE_FEDORA="fedora:36"
IMAGE_AMAZON="amazonlinux:latest"
REPO_RHEL="https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo"
REPO_AMAZON="https://rpm.releases.hashicorp.com/AmazonLinux/hashicorp.repo"
ARCH=( "386" "amd64" "arm64" "armv6" )

# set this so we can locate and execute the individual verification scripts.
SCRIPT_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# Arguments:
#   $1 - product (one of consul or consul-enterprise)
#   $2 - distro (one of ubuntu, debian, centos fedora or amazon)
#   $3 - version to match against (eg. 1.13.0-dev, 1.12.2)
function main {
  local product="${1:-}"
  local distro="${2:-}"
  local version="${3:-}"
  local image
  local script
  local script_args=()

  if [[ -z "${product}" ]]; then
    echo "ERROR: product argument is required"
    usage
    exit 1
  fi

  if [[ -z "${distro}" ]]; then
    echo "ERROR: distro argument is required"
    usage
    exit 1
  fi

  if [[ -z "${version}" ]]; then
    echo "ERROR: version argument is required"
    usage
    exit 1
  fi

  if [[ ! "${product}" = $PRODUCT_OSS ]] && [[ ! "${product}" = $PRODUCT_ENT ]]; then
    echo "ERROR: product must be one of $PRODUCT_OSS, $PRODUCT_ENT, got ${product}"
    usage
    exit 1
  fi

  script_args+=( "${product}" "${version}" )
  case "${distro}" in
    "ubuntu") 
      image=$IMAGE_UBUNTU
      script="apt_info.sh"
      ;;
    "debian")
      image=$IMAGE_DEBIAN
      script="apt_info.sh"
      ;;
    "centos")
      image=$IMAGE_CENTOS
      script="yum_info.sh"
      script_args+=( "${REPO_RHEL}" )
      ;;
    "fedora")
      image=$IMAGE_FEDORA
      script="dnf_info.sh"
      ;;
    "amazon")
      image=$IMAGE_AMAZON
      script="yum_info.sh"
      script_args+=( "${REPO_AMAZON}" )
      ;;
    *)
      echo "distro must be one of ubuntu, debian, centos, fedora, amazon; got ${distro}"
      exit 1
  esac
  
  # for arch in ${ARCH[@]}; do
    # echo "verifying ${distro} for ${arch}"
    docker pull "${image}"
    docker run -it -v ${SCRIPT_DIR}:/scripts "${image}" "/scripts/${script}" "${script_args[@]}"
    # docker run --platform $(platform_for_arch "${arch}") -it -v ${SCRIPT_DIR}:/scripts "${image}" "/scripts/${script}" "${script_args[@]}"
  # done
}

function platform_for_arch {
  local arch="${1:-}"
  case "$arch" in
    "386")
      echo "linux/386"
      ;;

    "amd64")
      echo "linux/amd64"
      ;;

    "arm64")
      echo "linux/arm64"
      ;;

    "armv6")
      echo "linux/arm/v6"
      ;;
  esac
}

main "$@"
