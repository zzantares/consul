package structs

import (
	"fmt"

	"github.com/hashicorp/consul/acl"
)

// ExternalServiceConfigEntry manages the configuration for an Envoy patch sets
// with the given name.
type ExternalServiceConfigEntry struct {
	// Kind of the config entry. This should be set to api.ExternalService.
	Kind string

	// Name is used to identify the patch set.
	Name string

	Type ExternalServiceConfigEntryType

	AWSLambda ExternalServiceConfigEntryAWSLambda

	Meta map[string]string `json:",omitempty"`

	EnterpriseMeta `hcl:",squash" mapstructure:",squash"` // formerly DestinationNS
	RaftIndex
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

func (e *ExternalServiceConfigEntry) GetKind() string {
	return ExternalService
}

func (e *ExternalServiceConfigEntry) GetName() string {
	if e == nil {
		return ""
	}

	return e.Name
}

func (e *ExternalServiceConfigEntry) GetMeta() map[string]string {
	if e == nil {
		return nil
	}
	return e.Meta
}

func (e *ExternalServiceConfigEntry) GetEnterpriseMeta() *EnterpriseMeta {
	if e == nil {
		return nil
	}

	return &e.EnterpriseMeta
}

func (e *ExternalServiceConfigEntry) Normalize() error {
	if e == nil {
		return fmt.Errorf("config entry is nil")
	}

	e.Kind = ExternalService
	e.EnterpriseMeta.Normalize()

	return nil
}

func (e *ExternalServiceConfigEntry) CanRead(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.ServiceRead(e.Name, &authzContext) == acl.Allow
}

func (e *ExternalServiceConfigEntry) CanWrite(authz acl.Authorizer) bool {
	var authzContext acl.AuthorizerContext
	e.FillAuthzContext(&authzContext)
	return authz.MeshWrite(&authzContext) == acl.Allow
}

func (e *ExternalServiceConfigEntry) GetRaftIndex() *RaftIndex {
	if e == nil {
		return &RaftIndex{}
	}

	return &e.RaftIndex
}

// TODO test and implement thisthis.
func (e *ExternalServiceConfigEntry) Validate() error {
	if err := validateConfigEntryMeta(e.Meta); err != nil {
		return err
	}

	return nil
}

func (e *ExternalServiceConfigEntry) ServiceName() ServiceName {
	return ServiceName{
		Name: e.Name,
	}
}
