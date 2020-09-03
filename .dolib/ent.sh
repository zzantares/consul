#!/usr/bin/env bash

# tell shellcheck about help
declare -A help; export help

help[ci-merge-oss-to-ent]="Start a CI job to merge OSS into Enterprise.

Args:
  branch - the branch to merge, defaults to 'master'

Environment Variables:
  CIRCLECI_CLI_TOKEN - a personal API token used to make the API request. Defaults
                       to the value in ~/.circleci/cli.yml. Create one at 
                       https://app.circleci.com/settings/user/tokens
"
ci-merge-oss-to-ent() {
    local default_token; default_token=$(grep 'token' ~/.circleci/cli.yml | awk '{print $2}')
    local token=${CIRCLECI_CLI_TOKEN-$default_token}
    local branch=${1-master}
    curl -sL \
        -X POST \
        --header "Circle-Token: ${token}" \
        --header "Content-Type: application/json" \
        -d "{\"build_parameters\": {\"CIRCLE_JOB\": \"oss-merge\", \"BRANCH\": \"${branch}\"}}" \
        'https://circleci.com/api/v1.1/project/github/hashicorp/consul-enterprise/tree/master' | \
    jq --raw-output '.build_url'
}
