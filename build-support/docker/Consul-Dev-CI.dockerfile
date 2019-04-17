# Note this is the same as the final stage of Consul-Dev-CI.dockerfile and
# should be kept roughly in sync.
FROM consul:latest

ADD /go/bin/consul /bin
