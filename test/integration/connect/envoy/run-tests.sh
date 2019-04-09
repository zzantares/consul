#!/bin/bash

#set -x
set -euo pipefail

DIR=$(cd -P -- "$(dirname -- "$0")" && pwd -P)

cd $DIR

# Start Consul first
docker-compose up -d consul

mkdir -p etc/{consul,envoy,bats}

source helpers.bash

RESULT=1

for c in "./case-*/" ; do
  echo "==> $( basename $c )"

  # Wipe state
  rm -rf etc/consul/*
  rm -rf etc/envoy/*
  rm -rf etc/bats/*

  # Reload consul config from defaults
  cp consul-base-cfg/* etc/consul

  # Add any overrides if there are any (no op if not)
  cp -f $DIR/${c}*.hcl $DIR/etc/consul 2>/dev/null || :

  # Reload consul
  retry 5 1 curl -X PUT localhost:8500/v1/agent/reload

  # Run test case setup (e.g. generating Envoy bootstrap)
  source ${c}setup.sh

  # Copy all the test files
  cp ${c}*.bats etc/bats
  cp helpers.bash etc/bats

  # Start services and proxies (TODO maybe make the list of services be
  # configured by test case?)
  docker-compose up -d s1 s1-sidecar-proxy s2 s2-sidecar-proxy

  # Execute tests
  if docker-compose up --abort-on-container-exit --exit-code-from verify verify ; then
    echo -n "==> $( basename $c ): "
    echogreen "✓ PASS"
  else
    echo -n "==> $( basename $c ): "
    echored "⨯ FAIL"
    if [ $RESULT -eq 1 ] ; then
      RESULT=0
    fi
  fi

  # Tear down services
  docker-compose stop s1 s1-sidecar-proxy s2 s2-sidecar-proxy
done

#docker-compose down

if [ $RESULT -eq 1 ] ; then
  echogreen "✓ PASS"
else
  echored "⨯ FAIL"
  exit 1
fi
