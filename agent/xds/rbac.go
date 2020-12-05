package xds

import (
	envoylistener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoyhttp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoyrbac "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	"github.com/hashicorp/consul/agent/structs"
)

func makeRBACNetworkFilter(intentions structs.Intentions, intentionDefaultAllow bool) (*envoylistener.Filter, error) {
	return nil, nil
}

func makeRBACHTTPFilter(intentions structs.Intentions, intentionDefaultAllow bool) (*envoyhttp.HttpFilter, error) {
	return nil, nil
}

func intentionListToIntermediateRBACForm(intentions structs.Intentions, isHTTP bool) []*rbacIntention {
	return nil
}

func removeSourcePrecedence(rbacIxns []*rbacIntention, intentionDefaultAction intentionAction) []*rbacIntention {
	return nil
}

func removeIntentionPrecedence(rbacIxns []*rbacIntention, intentionDefaultAction intentionAction) []*rbacIntention {
	return nil
}

func removePermissionPrecedence(perms []*rbacPermission, intentionDefaultAction intentionAction) []*rbacPermission {
	return nil
}

func intentionToIntermediateRBACForm(ixn *structs.Intention, isHTTP bool) *rbacIntention {
	return nil
}

type intentionAction int

const (
	intentionActionDeny intentionAction = iota
	intentionActionAllow
	intentionActionLayer7
)

func intentionActionFromBool(v bool) intentionAction {
	return 0
}
func intentionActionFromString(s structs.IntentionAction) intentionAction {
	return 0
}

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

func simplifyNotSourceSlice(notSources []structs.ServiceName) []structs.ServiceName {
	return nil
}

// makeRBACRules translates Consul intentions into RBAC Policies for Envoy.
//
// Consul lets you define up to 9 different kinds of intentions that apply at
// different levels of precedence (this is limited to 4 if not using Consul
// Enterprise). Each intention in this flat list (sorted by precedence) can either
// be an allow rule or a deny rule. Here’s a concrete example of this at work:
//
//     intern/trusted-app => billing/payment-svc : ALLOW (prec=9)
//     intern/*           => billing/payment-svc : DENY  (prec=8)
//     */*                => billing/payment-svc : ALLOW (prec=7)
//     ::: ACL default policy :::                : DENY  (prec=N/A)
//
// In contrast, Envoy lets you either configure a filter to be based on an
// allow-list or a deny-list based on the action attribute of the RBAC rules
// struct.
//
// On the surface it would seem that the configuration model of Consul
// intentions is incompatible with that of Envoy’s RBAC engine. For any given
// destination service Consul’s model requires evaluating a list of rules and
// short circuiting later rules once an earlier rule matches. After a rule is
// found to match then we decide if it is allow/deny. Envoy on the other hand
// requires the rules to express all conditions to allow access or all conditions
// to deny access.
//
// Despite the surface incompatibility it is possible to marry these two
// models. For clarity I’ll rewrite the earlier example intentions in an
// abbreviated form:
//
//     A         : ALLOW
//     B         : DENY
//     C         : ALLOW
//     <default> : DENY
//
// 1. Given that the overall intention default is set to deny, we start by
//    choosing to build an allow-list in Envoy (this is also the variant that I find
//    easier to think about).
// 2. Next we traverse the list in precedence order (top down) and any DENY
//    intentions are combined with later intentions using logical operations.
// 3. Now that all of the intentions result in the same action (allow) we have
//    successfully removed precedence and we can express this in as a set of Envoy
//    RBAC policies.
//
// After this the earlier A/B/C/default list becomes:
//
//     A            : ALLOW
//     C AND NOT(B) : ALLOW
//     <default>    : DENY
//
// Which really is just an allow-list of [A, C AND NOT(B)]
func makeRBACRules(intentions structs.Intentions, intentionDefaultAllow bool, isHTTP bool) (*envoyrbac.RBAC, error) {
	return nil, nil
}

func removeSameSourceIntentions(intentions structs.Intentions) structs.Intentions {
	return nil
}

// ixnSourceMatches deterines if the 'tester' service name is matched by the
// 'against' service name via wildcard rules.
//
// For instance:
// - (web, api)               => false, because these have no wildcards
// - (web, *)                 => true,  because "all services" includes "web"
// - (default/web, default/*) => true,  because "all services in the default NS" includes "default/web"
// - (default/*, */*)         => true,  "any service in any NS" includes "all services in the default NS"
func ixnSourceMatches(tester, against structs.ServiceName) bool {
	return false
}

// countWild counts the number of wildcard values in the given namespace and name.
func countWild(src structs.ServiceName) int {

	return 0
}

func andPrincipals(ids []*envoyrbac.Principal) *envoyrbac.Principal {
	return nil
}

func notPrincipal(id *envoyrbac.Principal) *envoyrbac.Principal {
	return nil
}

func idPrincipal(src structs.ServiceName) *envoyrbac.Principal {
	return nil
}
func makeSpiffePattern(sourceNS, sourceName string) string {
	return ""
}

func anyPermission() *envoyrbac.Permission {
	return nil
}

func convertPermission(perm *structs.IntentionPermission) *envoyrbac.Permission {
	return nil
}

func notPermission(perm *envoyrbac.Permission) *envoyrbac.Permission {
	return nil
}

func andPermissions(perms []*envoyrbac.Permission) *envoyrbac.Permission {
	return nil
}
