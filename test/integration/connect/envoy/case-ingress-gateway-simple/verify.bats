#!/usr/bin/env bats

load helpers

@test "ingress proxy admin is up on :20000" {
  retry_default curl -f -s localhost:20000/stats -o /dev/null
}

@test "s1 proxy admin is up on :19000" {
  retry_default curl -f -s localhost:19000/stats -o /dev/null
}

@test "s2 proxy admin is up on :19003" {
  retry_default curl -f -s localhost:19003/stats -o /dev/null
}

@test "ingress should be able to connect to s1 via configured port" {
}
