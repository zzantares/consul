// +build !tiny

package agent

import (
	"strings"

	"github.com/hashicorp/go-discover"
	discoverk8s "github.com/hashicorp/go-discover/provider/k8s"
	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/consul/lib"
)

func newDiscover() (*discover.Discover, error) {
	providers := make(map[string]discover.Provider)
	for k, v := range discover.Providers {
		providers[k] = v
	}
	providers["k8s"] = &discoverk8s.Provider{}

	return discover.New(
		discover.WithUserAgent(lib.UserAgent()),
		discover.WithProviders(providers),
	)
}

func retryJoinAddrs(disco *discover.Discover, variant, cluster string, retryJoin []string, logger hclog.Logger) []string {
	var addrs []string
	for _, addr := range retryJoin {
		switch {
		case strings.Contains(addr, "provider="):
			servers, err := disco.Addrs(addr, logger.StandardLogger(&hclog.StandardLoggerOptions{
				InferLevels: true,
			}))
			if err != nil {
				if logger != nil {
					logger.Error("Cannot discover address",
						"address", addr,
						"error", err,
					)
				}
			} else {
				addrs = append(addrs, servers...)
				if logger != nil {
					if variant == retryJoinMeshGatewayVariant {
						logger.Info("Discovered mesh gateways",
							"cluster", cluster,
							"mesh_gateways", strings.Join(servers, " "),
						)
					} else {
						logger.Info("Discovered servers",
							"cluster", cluster,
							"servers", strings.Join(servers, " "),
						)
					}
				}
			}

		default:
			addrs = append(addrs, addr)
		}
	}

	return addrs
}
