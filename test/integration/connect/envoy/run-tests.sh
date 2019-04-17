#!/bin/bash

set -euo pipefail

# DEBUG=1 enabled -x for this script so echos every command run
DEBUG=${DEBUG:-}

# FILTER_TESTS="<pattern>" skips any test whose CASENAME doesn't match the
# pattern. CASENAME is combination of the name from the case-<name> dir and the
# envoy version for example: "http, envoy 1.8.0". The pattern is passed to grep
# over that string.
FILTER_TESTS=${FILTER_TESTS:-}

# LEAVE_CONSUL_UP=1 leaves the consul container running at the end which can be
# useful for debugging.
LEAVE_CONSUL_UP=${LEAVE_CONSUL_UP:-}

# PROXY_LOGS_ON_FAIL=1 can be used in teardown scripts to dump logs from Envoy
# containers before stopping them. This can be useful for debugging but is very
# verbose so not done by default on a fail. See example in
# case-statsd-udp/teardown.sh for how to use it.
PROXY_LOGS_ON_FAIL=${PROXY_LOGS_ON_FAIL:-}

if [ ! -z "$DEBUG" ] ; then
  set -x
fi

DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)

cd $DIR

ENVOY_VERSIONS="1.8.0 1.9.1"

FILTER_TESTS=${FILTER_TESTS:-}
LEAVE_CONSUL_UP=${LEAVE_CONSUL_UP:-}
PROXY_LOGS_ON_FAIL=${PROXY_LOGS_ON_FAIL:-}

mkdir -p etc/{consul,envoy,bats,statsd}

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
    rm -rf etc/statsd/*

    # Reload consul config from defaults
    cp consul-base-cfg/* etc/consul

    # Add any overrides if there are any (no op if not)
    cp -f $DIR/${c}*.hcl $DIR/etc/consul 2>/dev/null || :

    # Start Consul first we do this here even though typically nothing stopped
    # it because it sometimes seems to be killed by something else (OOM killer)?
    docker-compose up -d consul

    # Reload consul
    echo "Reloading Consul config"
    if ! retry 10 2 docker_consul reload ; then
      # Clean up everything before we abort
      #docker-compose down
      echored "⨯ FAIL - couldn't reload consul config"
      exit 1
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
