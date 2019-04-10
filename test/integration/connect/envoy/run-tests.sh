#!/bin/bash

#set -x
set -euo pipefail

DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)

cd $DIR

ENVOY_VERSIONS="1.8.0 1.9.1"

FILTER_TESTS=${FILTER_TESTS:-}
LEAVE_CONSUL_UP=${LEAVE_CONSUL_UP:-}

# Start Consul first
docker-compose up -d consul

mkdir -p etc/{consul,envoy,bats}

source helpers.bash

RESULT=1

for c in ./case-*/ ; do
  for ev in $ENVOY_VERSIONS ; do
    CASENAME="$( basename $c | cut -c6- ), envoy $ev"
    echo ""
    echo "==> CASE $CASENAME"

    export ENVOY_VERSION=$ev

    if [ ! -z "$FILTER_TESTS" ] && echo "$CASENAME" | grep -v "$FILTER_TESTS" > /dev/null ; then
      echo "   SKIPPED: doesn't match FILTER_TESTS=$FILTER_TESTS"
      continue 1
    fi

    # Wipe state
    rm -rf etc/consul/*
    rm -rf etc/envoy/*
    rm -rf etc/bats/*

    # Reload consul config from defaults
    cp consul-base-cfg/* etc/consul

    # Add any overrides if there are any (no op if not)
    cp -f $DIR/${c}*.hcl $DIR/etc/consul 2>/dev/null || :

    # Reload consul
    echo "Reloading Consul config"
    if [ ! retry 10 2 curl -X PUT localhost:8500/v1/agent/reload] ; then
      # Clean up everything before we abort
      docker-compose down
    fi

    # Copy all the test files
    cp ${c}*.bats etc/bats
    cp helpers.bash etc/bats

    # Run test case setup (e.g. generating Envoy bootstrap, starting containers)
    source ${c}setup.sh

    # Execute tests
    if docker-compose up --build --abort-on-container-exit --exit-code-from verify verify ; then
      echo -n "==> CASE $CASENAME: "
      echogreen "✓ PASS"
    else
      echo -n "==> CASE $CASENAME: "
      echored "⨯ FAIL"
      if [ $RESULT -eq 1 ] ; then
        RESULT=0
      fi
    fi

    # Run test case teardown
    source ${c}teardown.sh
  done
done

if [ -z "$LEAVE_CONSUL_UP" ] ; then
  docker-compose down
fi

if [ $RESULT -eq 1 ] ; then
  echogreen "✓ PASS"
else
  echored "⨯ FAIL"
  exit 1
fi
