data_dir = "/tmp/data"
log_level = "TRACE"
datacenter = "dc1"
primary_datacenter = "dc1"
server = true
bootstrap_expect = 1
ui = true
bind_addr = "0.0.0.0"
client_addr = "0.0.0.0"
ports {
  grpc = 8502
}
connect {
  enabled = true
}
rpc {
  enable_streaming = true
}
use_streaming_backend = true
enable_central_service_config = true
