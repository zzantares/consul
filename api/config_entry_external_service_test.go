package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPI_ConfigEntries_ExternalService(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	config_entries := c.ConfigEntries()

	externalService1 := &ExternalServiceConfigEntry{
		Kind: ExternalService,
		Name: "foo",
		Meta: map[string]string{
			"foo": "bar",
			"gir": "zim",
		},
		Type: ExternalServiceConfigEntryTypeAWSLambda,
		AWSLambda: ExternalServiceConfigEntryAWSLambda{
			ARN:                "123",
			Region:             "234",
			PayloadPassthrough: false,
		},
	}

	_, wm, err := config_entries.Set(externalService1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// get it
	entry, qm, err := config_entries.Get(ExternalService, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)

	// verify it
	readExternalService, ok := entry.(*ExternalServiceConfigEntry)
	require.True(t, ok)
	require.Equal(t, externalService1.Kind, readExternalService.Kind)
	require.Equal(t, externalService1.Name, readExternalService.Name)
	require.Equal(t, externalService1.Type, readExternalService.Type)
	require.Equal(t, externalService1.Meta, readExternalService.Meta)
	require.Equal(t, externalService1.Meta, readExternalService.GetMeta())

	// update it
	externalService1.AWSLambda.Region = "new-region"

	// CAS fail
	written, _, err := config_entries.CAS(externalService1, 0, nil)
	require.NoError(t, err)
	require.False(t, written)

	// CAS success
	written, wm, err = config_entries.CAS(externalService1, readExternalService.ModifyIndex, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)
	require.True(t, written)

	// update no cas
	externalService1.AWSLambda.Region = "newer-region"

	_, wm, err = config_entries.Set(externalService1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// list them
	entries, qm, err := config_entries.List(ExternalService, nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)
	require.Len(t, entries, 1)

	entry = entries[0]
	// verify it
	readExternalService, ok = entry.(*ExternalServiceConfigEntry)
	require.True(t, ok)
	require.Equal(t, externalService1.Kind, readExternalService.Kind)
	require.Equal(t, externalService1.Type, readExternalService.Type)
	require.Equal(t, externalService1.Name, readExternalService.Name)
	require.Equal(t, externalService1.Meta, readExternalService.Meta)
	require.Equal(t, externalService1.AWSLambda, readExternalService.AWSLambda)
	require.Equal(t, externalService1.Meta, readExternalService.GetMeta())

	// delete it
	wm, err = config_entries.Delete(ExternalService, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// verify deletion
	_, _, err = config_entries.Get(ExternalService, "foo", nil)
	require.Error(t, err)
}
