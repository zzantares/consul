#!/bin/bash

set -eEuo pipefail

# wait for bootstrap to apply config entries
wait_for_config_entry ingress-gateway ingress-gateway

gen_envoy_bootstrap ingress-gateway 20000 primary true
gen_envoy_bootstrap s2 19001 primary