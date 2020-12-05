package xds

import (
	envoyrbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
)

type rbacPermission struct {
}

// TODO(m1) removing this line also eliminates the problem
func (p *rbacPermission) Flatten() *envoyrbac.Permission {
	return nil
}
