#!/usr/bin/env bats

load helpers

@test "s1 proxy admin is up on :19000" {
  retry 5 1 curl -f -s localhost:19000/stats -o /dev/null
}

@test "s2 proxy admin is up on :19001" {
  retry 5 1 curl -f -s localhost:19001/stats -o /dev/null
}

@test "s1 proxy listener shoudl be up and have right cert" {
  openssl s_client -connect localhost:21000 \
    -showcerts 2>/dev/null \
    | openssl x509 -noout -text \
    | grep -Eo 'URI:spiffe://([a-zA-Z0-9-]+).consul/ns/default/dc/dc1/svc/s1'
}

@test "s2 proxy listener shoudl be up and have right cert" {
  openssl s_client -connect localhost:21001 \
    -showcerts 2>/dev/null \
    | openssl x509 -noout -text \
    | grep -Eo 'URI:spiffe://([a-zA-Z0-9-]+).consul/ns/default/dc/dc1/svc/s2'
}

@test "s1 upstream should be able to connect to s2" {
  run curl -s -f -d hello localhost:5000
  [ "$status" -eq 0 ]
  [ "$output" = "hello" ]
}
