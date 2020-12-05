package xds

import (
	envoyrbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/hashicorp/consul/agent/structs"
)

type intentionAction int

const (
	intentionActionDeny intentionAction = iota
	intentionActionAllow
	intentionActionLayer7
)

type rbacIntention struct {
	Source      structs.ServiceName
	NotSources  []structs.ServiceName
	Action      intentionAction
	Permissions []*rbacPermission
	Precedence  int

	// Skip is field used to indicate that this intention can be deleted in the
	// final pass. Items marked as true should generally not escape the method
	// that marked them.
	Skip bool

	ComputedPrincipal *envoyrbac.Principal
}

func (r *rbacIntention) FlattenPrincipal() *envoyrbac.Principal {
	return nil
}

type rbacPermission struct {
	Definition *structs.IntentionPermission

	Action   intentionAction
	Perm     *envoyrbac.Permission
	NotPerms []*envoyrbac.Permission

	// Skip is field used to indicate that this permission can be deleted in
	// the final pass. Items marked as true should generally not escape the
	// method that marked them.
	Skip bool

	ComputedPermission *envoyrbac.Permission
}

func (p *rbacPermission) Flatten() *envoyrbac.Permission {
	return nil
}
