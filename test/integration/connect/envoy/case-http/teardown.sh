#!/bin/bash

set -euo pipefail

# Tear down services
docker-compose stop s1 s1-sidecar-proxy s2 s2-sidecar-proxy