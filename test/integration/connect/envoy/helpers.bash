#!/bin/bash

# retry based on
# https://github.com/fernandoacorreia/azure-docker-registry/blob/master/tools/scripts/create-registry-server
# under MIT license.
function retry {
  local n=1
  local max=$1
  shift
  local delay=$1
  shift
  while true; do
    "$@" && break || {
      exit=$?
      if [[ $n -lt $max ]]; then
        ((n++))
        echo "Command failed. Attempt $n/$max:"
        sleep $delay;
      else
        echo "The command has failed after $n attempts." >&2
        return $exit
      fi
    }
  done
}


function echored {
  tput setaf 1
  tput bold
  echo $@
  tput sgr0
}

function echogreen {
  tput setaf 2
  tput bold
  echo $@
  tput sgr0
}

function get_cert {
  local HOSTPORT=$1
  openssl s_client -connect $HOSTPORT \
    -showcerts 2>/dev/null \
    | openssl x509 -noout -text
}

function assert_proxy_presents_cert_uri {
  local HOSTPORT=$1
  local SERVICENAME=$2

  CERT=$(retry 5 1 get_cert $HOSTPORT)

  echo "WANT SERVICE: $SERVICENAME"
  echo "GOT CERT:"
  echo "$CERT"

  echo "$CERT" | grep -Eo "URI:spiffe://([a-zA-Z0-9-]+).consul/ns/default/dc/dc1/svc/$SERVICENAME"
}

function get_envoy_listener_filters {
  local HOSTPORT=$1
  run retry 5 1 curl -s -f $HOSTPORT/config_dump
  [ "$status" -eq 0 ]
  echo "$output" | jq --raw-output '.configs[2].dynamic_active_listeners[].listener | "\(.name) \( .filter_chains[0].filters | map(.name) | join(","))"'
}

function docker_consul {
  docker run -ti -v $(pwd):/var/wd -w /var/wd --network container:envoy_consul_1 consul-dev $@
}