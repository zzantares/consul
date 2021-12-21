package structs

import (
	"fmt"

	"github.com/hashicorp/consul/acl"
)

type PatchType string

const (
	Replace PatchType = "replace"
)

type ApplyTo string

const (
	ApplyToServiceFilter  ApplyTo = "service_filter"
	ApplyToServiceCluster ApplyTo = "service_cluster"
)

type PatchMode string

const (
	PatchModeConnectProxy       PatchMode = "connect_proxy"
	PatchModeTerminatingGateway PatchMode = "terminating_gateway"
)

type Patch struct {
	ApplyTo ApplyTo
	Mode    PatchMode
	Type    PatchType
	Path    string
	Value   string
}

// EnvoyPatchSetConfigEntry manages the configuration for an Envoy patch sets
// with the given name.
type EnvoyPatchSetConfigEntry struct {
	// Kind of the config entry. This should be set to api.EnvoyPatchSet.
	Kind string

	// Name is used to identify the patch set.
	Name string

	Version string

	// TODO why are namespace and partition not here?

	Meta map[string]string `json:",omitempty"`

	Patches []Patch

	EnterpriseMeta `hcl:",squash" mapstructure:",squash"` // formerly DestinationNS
	RaftIndex
}

func (e *EnvoyPatchSetConfigEntry) GetKind() string {
	return EnvoyPatchSet
}

func (e *EnvoyPatchSetConfigEntry) GetName() string {
	if e == nil {
		return ""
	}

	return e.Name
}

func (e *EnvoyPatchSetConfigEntry) GetMeta() map[string]string {
	if e == nil {
		return nil
	}
	return e.Meta
}

func (e *EnvoyPatchSetConfigEntry) GetEnterpriseMeta() *EnterpriseMeta {
	if e == nil {
		return nil
	}

	return &e.EnterpriseMeta
}

func (e *EnvoyPatchSetConfigEntry) Normalize() error {
	if e == nil {
		return fmt.Errorf("config entry is nil")
	}

	e.Kind = EnvoyPatchSet
	e.EnterpriseMeta.Normalize()

	return nil
}

func (e *EnvoyPatchSetConfigEntry) CanRead(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.ServiceRead(e.Name, &authzContext) == acl.Allow
}

func (e *EnvoyPatchSetConfigEntry) CanWrite(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.MeshWrite(&authzContext) == acl.Allow
}

func (e *EnvoyPatchSetConfigEntry) GetRaftIndex() *RaftIndex {
	if e == nil {
		return &RaftIndex{}
	}

	return &e.RaftIndex
}

// TODO test this.
func (e *EnvoyPatchSetConfigEntry) Validate() error {
	if err := validateConfigEntryMeta(e.Meta); err != nil {
		return err
	}

	if e.Name == "" {
		return fmt.Errorf("Service name cannot be blank.")
	}

	if e.Version == "" {
		return fmt.Errorf("Version cannot be blank.")
	}

	if len(e.Patches) == 0 {
		return fmt.Errorf("At least one patch must exist.")
	}

	// validate enterprise meta
	return nil
}

func (e *EnvoyPatchSetConfigEntry) GetEnvoyPatchSetIdentifier() ApplyEnvoyPatchSetIdentifier {
	return ApplyEnvoyPatchSetIdentifier{
		Name:    e.Name,
		Version: e.Version,
	}
}

// ApplyEnvoyPatchSetConfigEntry manages the configuration for apply Envoy patch sets
// with a given name.
type ApplyEnvoyPatchSetConfigEntry struct {
	// Kind of the config entry. This should be set to api.ApplyEnvoyPatchSet.
	Kind string

	// Name is used to identify the patch set.
	Name string

	EnvoyPatchSet ApplyEnvoyPatchSetIdentifier

	ApplyIndex int

	Filter ApplyEnvoyPatchSetFilter

	// TODO why are namespace and partition not here?

	Meta map[string]string `json:",omitempty"`

	Patches []Patch

	EnterpriseMeta `hcl:",squash" mapstructure:",squash"` // formerly DestinationNS
	RaftIndex
}

type ApplyEnvoyPatchSetIdentifier struct {
	Name    string
	Version string
}

type ApplyEnvoyPatchSetFilter struct {
	Service string
}

func (e *ApplyEnvoyPatchSetConfigEntry) GetKind() string {
	return ApplyEnvoyPatchSet
}

func (e *ApplyEnvoyPatchSetConfigEntry) GetName() string {
	if e == nil {
		return ""
	}

	return e.Name
}

func (e *ApplyEnvoyPatchSetConfigEntry) GetMeta() map[string]string {
	if e == nil {
		return nil
	}
	return e.Meta
}

func (e *ApplyEnvoyPatchSetConfigEntry) GetEnterpriseMeta() *EnterpriseMeta {
	if e == nil {
		return nil
	}

	return &e.EnterpriseMeta
}

func (e *ApplyEnvoyPatchSetConfigEntry) Normalize() error {
	if e == nil {
		return fmt.Errorf("config entry is nil")
	}

	e.Kind = ApplyEnvoyPatchSet
	e.EnterpriseMeta.Normalize()

	return nil
}

func (e *ApplyEnvoyPatchSetConfigEntry) CanRead(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.ServiceRead(e.Name, &authzContext) == acl.Allow
}

func (e *ApplyEnvoyPatchSetConfigEntry) CanWrite(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.MeshWrite(&authzContext) == acl.Allow
}

func (e *ApplyEnvoyPatchSetConfigEntry) GetRaftIndex() *RaftIndex {
	if e == nil {
		return &RaftIndex{}
	}

	return &e.RaftIndex
}

// TODO test this.
func (e *ApplyEnvoyPatchSetConfigEntry) Validate() error {
	if err := validateConfigEntryMeta(e.Meta); err != nil {
		return err
	}

	if e.Name == "" {
		return fmt.Errorf("Service name cannot be blank.")
	}

	if e.EnvoyPatchSet.Name == "" {
		return fmt.Errorf("Patch set application Name cannot be blank.")
	}

	if e.EnvoyPatchSet.Version == "" {
		return fmt.Errorf("Patch set application Version cannot be blank.")
	}

	if e.Filter.Service == "" && e.ApplyIndex == 0 {
		return fmt.Errorf("Either the filter service or apply index must be populated")
	}

	if e.Filter.Service != "" && e.ApplyIndex > 0 {
		return fmt.Errorf("Both the filter service or apply index can't be populated")
	}

	// validate enterprise meta
	return nil
}
