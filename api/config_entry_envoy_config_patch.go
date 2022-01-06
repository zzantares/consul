package api

type Patch struct {
	Mode   PatchMode
	Entity Entity
	Type   PatchType
	Path   string // Should just straight up work with reflection.
	Value  string

	// Can use reflection to use json patch path syntax:
	// {
	//   "/filterChains/0/filters/name": "envoy.filters.network.tcp_proxy"
	// }
	// This gives us the ability to do arbitrary matching without introducing a
	// complicated new syntax.
	PathMatches map[string]interface{}
	// TODO add schema. This isn't strictly necessary, but it would result in a better user
	// experience and simpler downstream code. What is the best way to do this in Go?
}

type PatchType string

const (
	Replace PatchType = "replace"
)

type Entity string

const (
	EntityFilter  Entity = "filter"
	EntityCluster Entity = "cluster"
)

type PatchMode string

const (
	PatchModeConnectProxy       PatchMode = "connect_proxy"
	PatchModeTerminatingGateway PatchMode = "terminating_gateway"
)

// EnvoyPatchSetConfigEntry manages the configuration for an Envoy patch sets
// with the given name.
type EnvoyPatchSetConfigEntry struct {
	// Kind of the config entry. This should be set to api.EnvoyPatchSet.
	Kind string

	// Either Service or Patch.PathMatches should be populated.
	Service bool

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

	Service string

	// Partition is the partition the IngressGateway is associated with.
	// Partitioning is a Consul Enterprise feature.
	Partition string `json:",omitempty"`

	// Namespace is the namespace the IngressGateway is associated with.
	// Namespacing is a Consul Enterprise feature.
	Namespace string `json:",omitempty"`

	Meta map[string]string `json:",omitempty"`

	// Eventually this will match the schema in the patch and be of type map[string]interface{}
	Arguments map[string]string `json:",omitempty"`

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

func (i *ApplyEnvoyPatchSetConfigEntry) GetKind() string            { return i.Kind }
func (i *ApplyEnvoyPatchSetConfigEntry) GetName() string            { return i.Name }
func (i *ApplyEnvoyPatchSetConfigEntry) GetPartition() string       { return i.Partition }
func (i *ApplyEnvoyPatchSetConfigEntry) GetNamespace() string       { return i.Namespace }
func (i *ApplyEnvoyPatchSetConfigEntry) GetMeta() map[string]string { return i.Meta }
func (i *ApplyEnvoyPatchSetConfigEntry) GetCreateIndex() uint64     { return i.CreateIndex }
func (i *ApplyEnvoyPatchSetConfigEntry) GetModifyIndex() uint64     { return i.ModifyIndex }
