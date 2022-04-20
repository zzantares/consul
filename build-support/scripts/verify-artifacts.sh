#!/bin/bash

set -euo pipefail

RELEASES_HOME=/tmp/hc-releases
ARTIFACTS_HOME=/tmp/hc-artifacts

RELEASE_BRANCH_PREFIX=release/

function usage {
    echo "usage:"
    echo "verify-artifacts.sh <repo-name=(consul|consul-enterprise)> <version> <channel=(dev|staging|production)>"
    echo
    echo "example:"
    echo "verify-artifacts.sh consul 1.12.0 dev"
}

function main {
    local repo_name
    local version
    local channel 
    repo_name="${1:-}"
    version="${2:-}"
    channel="${3:-}"

    if [[ -z "${repo_name}" ]]; then
        echo "error: repo_name argument is required"
        usage
        exit 1
    fi

    if [[ -z "${version}" ]]; then
        echo "error: version argument is required"
        usage
        exit 1
    fi

    if [[ -z "${channel}" ]]; then
        echo "error: channel argument is required"
        usage
        exit 1
    fi

    # TODO:
    # validate repo_name is (consul|consul-enterprise)
    # validate channel is (dev|staging|production)
    # check that bob CLI is installed
    # check that awscli is installed

    local repo_path="${RELEASES_HOME}/${repo_name}"
    local branch_name="${RELEASE_BRANCH_PREFIX}${version}"

    # delete the repo if it exists so that we can ensure we get the latest ref
    if [[ -d "${repo_path}" ]]; then
        rm -rf "${repo_path}"
    fi
    # shallow clone at the release branch so that we can get the SHA
    git clone \
        --depth 1 \
        --quiet \
        git@github.com:hashicorp/${repo_name} \
        -b "${branch_name}" \
        "${repo_path}"

    # get SHA for release branch
    pushd "${repo_path}"
    local sha=$(git rev-parse HEAD)
    local output_dir="$ARTIFACTS_HOME/${channel}"
    mkdir -p "${output_dir}" || true

    local product_version="${version}"
    if [[ "${repo_name}" == "consul-enterprise" ]]; then
        product_version="${version}+ent"
    fi

    # download all artifacts for darwin, linux and docker (docker matches 'linux' pattern)
    for os in darwin linux; do
        echo "Downloading ${repo_name} ${version} ${sha} for ${os}"
        bob download artifactory \
            -channel "${channel}" \
            -commit="${sha}" \
            -product-name="${repo_name}" \
            -product-version="${product_version}" \
            -pattern '*'$os'*' \
            -output-dir="${output_dir}"
    done
    # if channel == staging
    #   auth with AWS
    #   aws s3 cp ...

    # if channel == production
    #   download from releases.hashicorp.com
    #   unzip/verify
    
    # for binaries from zips
    for f in $(find "${output_dir}" -name '*.zip' -print); do
        bin=${f%.*}
        unzip ${f} && mv consul ${bin}
        rm ${f}

        # smoke test
        # echo "Executing $bin version"
        # ./$bin version
        # echo "--------"
        # echo;
    done

    # for binaries from tarballs
    # untar artifact
    # try to execute smoke test (in container?)

    # for docker images
    # if staging, need to generate the URI for artifactory

    popd
}

main "$@"
