#!/bin/bash

set -euo pipefail

# Tear down services
docker-compose stop s1-sidecar-proxy-consul-exec