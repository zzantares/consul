data_dir = "/tmp/data"
log_level = "TRACE"
datacenter ="dc1"
primary_datacenter = "dc1"
use_streaming_backend = true
enable_central_service_config = true

bind_addr = "0.0.0.0"
client_addr = "0.0.0.0"

check_update_interval = "0s"
ui = true

http_config {
	use_cache = true
}
dns_config {
	use_cache  =true
}
ports {
  grpc = 8502
}

retry_join = ["server1"]
