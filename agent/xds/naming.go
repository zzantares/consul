package xds

import (
	"github.com/hashicorp/consul/agent/structs"
)

func CustomizeClusterName(clusterName string, chain *structs.CompiledDiscoveryChain) string {
	return ""
}
