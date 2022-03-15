package consul

import (
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/grpc/public/services/dataplane"
	"github.com/hashicorp/consul/agent/structs"
)

type DataplaneBackend struct {
	Srv *Server
}

func (s DataplaneBackend) ResolveTokenAndDefaultMeta(
	token string,
	entMeta *structs.EnterpriseMeta,
	authzContext *acl.AuthorizerContext,
) (acl.Authorizer, error) {
	return s.Srv.ResolveTokenAndDefaultMeta(token, entMeta, authzContext)
}

var _ dataplane.Backend = (*DataplaneBackend)(nil)
