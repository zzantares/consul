//go:build !consulent
// +build !consulent

package consul

import "github.com/hashicorp/consul/agent/structs"

func (s *Server) replicationEnterpriseMeta() *structs.EnterpriseMeta {
	return structs.ReplicationEnterpriseMeta()
}
