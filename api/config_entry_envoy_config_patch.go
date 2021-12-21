package api

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

	// Partition is the partition the IngressGateway is associated with.
	// Partitioning is a Consul Enterprise feature.
	Partition string `json:",omitempty"`

	// Namespace is the namespace the IngressGateway is associated with.
	// Namespacing is a Consul Enterprise feature.
	Namespace string `json:",omitempty"`

	Meta map[string]string `json:",omitempty"`

	// CreateIndex is the Raft index this entry was created at. This is a
	// read-only field.
	CreateIndex uint64

	// ModifyIndex is used for the Check-And-Set operations and can also be fed
	// back into the WaitIndex of the QueryOptions in order to perform blocking
	// queries.
	ModifyIndex uint64

	Patches []Patch
}

func (i *EnvoyPatchSetConfigEntry) GetKind() string            { return i.Kind }
func (i *EnvoyPatchSetConfigEntry) GetName() string            { return i.Name }
func (i *EnvoyPatchSetConfigEntry) GetPartition() string       { return i.Partition }
func (i *EnvoyPatchSetConfigEntry) GetNamespace() string       { return i.Namespace }
func (i *EnvoyPatchSetConfigEntry) GetMeta() map[string]string { return i.Meta }
func (i *EnvoyPatchSetConfigEntry) GetCreateIndex() uint64     { return i.CreateIndex }
func (i *EnvoyPatchSetConfigEntry) GetModifyIndex() uint64     { return i.ModifyIndex }

// EnvoyPatchSetConfigEntry manages the application of Envoy patch sets
// with the given name.
type ApplyEnvoyPatchSetConfigEntry struct {
	// Kind of the config entry. This should be set to api.ApplyEnvoyPatchSet.
	Kind string

	// Name is used to identify the patch set.
	Name string

	EnvoyPatchSet ApplyEnvoyPatchSetIdentifier

	ApplyIndex int

	Filter ApplyEnvoyPatchSetFilter

	// Partition is the partition the IngressGateway is associated with.
	// Partitioning is a Consul Enterprise feature.
	Partition string `json:",omitempty"`

	// Namespace is the namespace the IngressGateway is associated with.
	// Namespacing is a Consul Enterprise feature.
	Namespace string `json:",omitempty"`

	Meta map[string]string `json:",omitempty"`

	// CreateIndex is the Raft index this entry was created at. This is a
	// read-only field.
	CreateIndex uint64

	// ModifyIndex is used for the Check-And-Set operations and can also be fed
	// back into the WaitIndex of the QueryOptions in order to perform blocking
	// queries.
	ModifyIndex uint64
}

type ApplyEnvoyPatchSetIdentifier struct {
	Name    string
	Version string
}

type ApplyEnvoyPatchSetFilter struct {
	Service string
}

func (i *ApplyEnvoyPatchSetConfigEntry) GetKind() string            { return i.Kind }
func (i *ApplyEnvoyPatchSetConfigEntry) GetName() string            { return i.Name }
func (i *ApplyEnvoyPatchSetConfigEntry) GetPartition() string       { return i.Partition }
func (i *ApplyEnvoyPatchSetConfigEntry) GetNamespace() string       { return i.Namespace }
func (i *ApplyEnvoyPatchSetConfigEntry) GetMeta() map[string]string { return i.Meta }
func (i *ApplyEnvoyPatchSetConfigEntry) GetCreateIndex() uint64     { return i.CreateIndex }
func (i *ApplyEnvoyPatchSetConfigEntry) GetModifyIndex() uint64     { return i.ModifyIndex }
