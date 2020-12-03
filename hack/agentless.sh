#!/bin/bash

set -euo pipefail

unset CDPATH

cd "$(dirname "$0")"

# readonly public_ip="$(ip -4 addr show docker0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}')"
# echo "Using public ip: ${public_ip}"

if ! consul info >/dev/null 2>&1 ; then
    die "consul not running; please run it as 'consul agent -dev'"
fi

# docker rm -f ping || true
# docker rm -f pong || true

# docker run -d --name ping --init -p 8081:8081 rboyer/pingpong \
#     -bind 0.0.0.0:8081 -dial pong:5000 -dialfreq 5ms -name ping
# docker run -d --name pong --init -p 9091:9091 rboyer/pingpong \
#     -bind 0.0.0.0:9091 -dial ping:6000 -dialfreq 5ms -name pong

killall pingpong || true

./pingpong -bind 127.0.0.1:8081 -dial 127.0.0.1:5000 -dialfreq 5ms -name ping &
./pingpong -bind 127.0.0.1:9091 -dial 127.0.0.1:6000 -dialfreq 5ms -name pong &

curl -sL -XPUT localhost:8500/v1/catalog/register -d'
{
  "Node": "box-ping",
  "Address": "127.0.0.1",
  "Service": {
    "Service": "ping",
    "Port": 8080,
    "Meta": {
      "connect_proxy": "true",
      "connect_local_bind": "127.0.0.1:8081",
      "connect_upstreams": "pong:5000"
    },
    "Connect": {
      "Native": true
    }
  }
}
'

curl -sL -XPUT localhost:8500/v1/catalog/register -d'
{
  "Node": "box-pong",
  "Address": "127.0.0.1",
  "Service": {
    "Service": "pong",
    "Port": 9090,
    "Meta": {
      "connect_proxy": "true",
      "connect_local_bind": "127.0.0.1:9091",
      "connect_upstreams": "ping:6000"
    },
    "Connect": {
      "Native": true
    }
  }
}
'

echo "================="
echo "term1: consul connect envoy -proxy-id box-ping/ping -agentless"
echo "term2: consul connect envoy -proxy-id box-pong/pong -agentless -admin-bind localhost:19001"
