#!/usr/bin/env bats

load helpers

@test "ingress proxy admin is up on :20000" {
  retry_default curl -f -s localhost:20000/stats -o /dev/null
}

@test "s2 proxy admin is up on :19001" {
  retry_default curl -f -s localhost:19001/stats -o /dev/null
}

@test "s2 proxy should be healthy" {
  assert_service_has_healthy_instances s2 1
}

@test "ingress-gateway should have healthy endpoints for s2" {
   assert_upstream_has_endpoints_in_status 127.0.0.1:20000 s2 HEALTHY 1
}

@test "s1 upstream should be able to connect to s2 via grpc" {
  run fortio grpcping localhost:9999

  echo "OUTPUT: $output"

  [ "$status" == 0 ]
}
