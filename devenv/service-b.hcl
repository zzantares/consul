{
  "service": {
    "name": "service-b",
    "port": 9091,
    "connect": {
      "sidecar_service": {
        "proxy": {
          "upstreams": [
            {
                "destination_name": "service-c",
                "local_bind_port": 9990
            }
          ]
        }
      }
    }
  }
}
