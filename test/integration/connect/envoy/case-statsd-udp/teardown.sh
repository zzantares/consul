#!/bin/bash

set -euo pipefail

# If the cast failed, dump the proxy logs
if [[ "$RESULT" == 0  && ! -z "$PROXY_LOGS_ON_FAIL" ]] ; then
  docker-compose logs s1-sidecar-proxy
  docker-compose logs s2-sidecar-proxy
  docker-compose logs fake-statsd
fi

# Tear down services
docker-compose stop s1 s1-sidecar-proxy s2 s2-sidecar-proxy