enable_central_service_config = true

config_entries {
  bootstrap {
    kind = "ingress-gateway"
    name = "ingress-gateway"

    listeners = [
      {
        port = 9999
        services = [
          {
            name = "s1"
          }
        ]
      }
    ]
  }
}
