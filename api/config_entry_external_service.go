package api

type ExternalServiceConfigEntry struct {
	// Kind of the config entry. This should be set to api.EnvoyPatchSet.
	Kind string

	// Name is used to identify the patch set.
	Name string

	Type ExternalServiceConfigEntryType

	AWSLambda ExternalServiceConfigEntryAWSLambda

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

type ExternalServiceConfigEntryType string

const (
	ExternalServiceConfigEntryTypeAWSLambda ExternalServiceConfigEntryType = "aws-lambda"
)

type ExternalServiceConfigEntryAWSLambda struct {
	ARN                string
	PayloadPassthrough bool
	Region             string
}

func (i *ExternalServiceConfigEntry) GetKind() string            { return i.Kind }
func (i *ExternalServiceConfigEntry) GetName() string            { return i.Name }
func (i *ExternalServiceConfigEntry) GetPartition() string       { return i.Partition }
func (i *ExternalServiceConfigEntry) GetNamespace() string       { return i.Namespace }
func (i *ExternalServiceConfigEntry) GetMeta() map[string]string { return i.Meta }
func (i *ExternalServiceConfigEntry) GetCreateIndex() uint64     { return i.CreateIndex }
func (i *ExternalServiceConfigEntry) GetModifyIndex() uint64     { return i.ModifyIndex }
