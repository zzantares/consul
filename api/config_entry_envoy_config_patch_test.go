package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPI_ConfigEntries_EnvoyPatchSet(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	config_entries := c.ConfigEntries()

	patchSet1 := &EnvoyPatchSetConfigEntry{
		Kind:    EnvoyPatchSet,
		Version: "0.0.1",
		Name:    "foo",
		Meta: map[string]string{
			"foo": "bar",
			"gir": "zim",
		},
		Service: true,
		Patches: []Patch{
			{
				Mode:   PatchModeTerminatingGateway,
				Entity: EntityFilter,
				Type:   Replace,
				Path:   "/",
				Value:  "{}",
			},
		},
	}

	_, wm, err := config_entries.Set(patchSet1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// get it
	entry, qm, err := config_entries.Get(EnvoyPatchSet, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)

	// verify it
	readPatchSet, ok := entry.(*EnvoyPatchSetConfigEntry)
	require.True(t, ok)
	require.Equal(t, patchSet1.Kind, readPatchSet.Kind)
	require.Equal(t, patchSet1.Name, readPatchSet.Name)
	require.Equal(t, patchSet1.Meta, readPatchSet.Meta)
	require.Equal(t, patchSet1.Meta, readPatchSet.GetMeta())

	// update it
	patchSet1.Patches = []Patch{
		{
			Type:  Replace,
			Path:  "/asdf",
			Value: "{}",
		},
	}

	// CAS fail
	written, _, err := config_entries.CAS(patchSet1, 0, nil)
	require.NoError(t, err)
	require.False(t, written)

	// CAS success
	written, wm, err = config_entries.CAS(patchSet1, readPatchSet.ModifyIndex, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)
	require.True(t, written)

	// update no cas
	patchSet1.Patches = []Patch{
		{
			Mode:   PatchModeConnectProxy,
			Entity: EntityFilter,
			Type:   Replace,
			Path:   "/",
			Value:  "not the same",
		},
	}
	_, wm, err = config_entries.Set(patchSet1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// list them
	entries, qm, err := config_entries.List(EnvoyPatchSet, nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)
	require.Len(t, entries, 1)

	entry = entries[0]
	// verify it
	readPatchSet, ok = entry.(*EnvoyPatchSetConfigEntry)
	require.True(t, ok)
	require.Equal(t, patchSet1.Kind, readPatchSet.Kind)
	require.Equal(t, patchSet1.Name, readPatchSet.Name)
	require.Equal(t, patchSet1.Meta, readPatchSet.Meta)
	require.Equal(t, patchSet1.Patches, readPatchSet.Patches)
	require.Equal(t, patchSet1.Meta, readPatchSet.GetMeta())

	// delete it
	wm, err = config_entries.Delete(EnvoyPatchSet, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// verify deletion
	_, _, err = config_entries.Get(EnvoyPatchSet, "foo", nil)
	require.Error(t, err)
}

func TestAPI_ConfigEntries_ApplyEnvoyPatchSet(t *testing.T) {
	t.Parallel()
	c, s := makeClient(t)
	defer s.Stop()

	config_entries := c.ConfigEntries()

	applyPatchSet1 := &ApplyEnvoyPatchSetConfigEntry{
		Kind: ApplyEnvoyPatchSet,
		Name: "foo",
		EnvoyPatchSet: ApplyEnvoyPatchSetIdentifier{
			Version: "0.0.1",
			Name:    "foo-patch-set",
		},
		Service: "service-name",
		Arguments: map[string]string{
			"foo": "bar",
			"gir": "zim",
		},
	}

	_, wm, err := config_entries.Set(applyPatchSet1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// get it
	entry, qm, err := config_entries.Get(ApplyEnvoyPatchSet, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)

	// verify it
	readPatchSet, ok := entry.(*ApplyEnvoyPatchSetConfigEntry)
	require.True(t, ok)
	require.Equal(t, applyPatchSet1.Kind, readPatchSet.Kind)
	require.Equal(t, applyPatchSet1.Name, readPatchSet.Name)
	require.Equal(t, applyPatchSet1.Meta, readPatchSet.Meta)
	require.Equal(t, applyPatchSet1.Arguments, readPatchSet.Arguments)

	// update it
	applyPatchSet1.EnvoyPatchSet.Version = "0.0.2"

	// CAS fail
	written, _, err := config_entries.CAS(applyPatchSet1, 0, nil)
	require.NoError(t, err)
	require.False(t, written)

	// CAS success
	written, wm, err = config_entries.CAS(applyPatchSet1, readPatchSet.ModifyIndex, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)
	require.True(t, written)

	// update no cas
	applyPatchSet1.EnvoyPatchSet.Version = "0.0.3"

	_, wm, err = config_entries.Set(applyPatchSet1, nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// list them
	entries, qm, err := config_entries.List(ApplyEnvoyPatchSet, nil)
	require.NoError(t, err)
	require.NotNil(t, qm)
	require.NotEqual(t, 0, qm.RequestTime)
	require.Len(t, entries, 1)

	entry = entries[0]
	// verify it
	readPatchSet, ok = entry.(*ApplyEnvoyPatchSetConfigEntry)
	require.True(t, ok)
	require.Equal(t, applyPatchSet1.Kind, readPatchSet.Kind)
	require.Equal(t, applyPatchSet1.Name, readPatchSet.Name)
	require.Equal(t, applyPatchSet1.Meta, readPatchSet.Meta)
	require.Equal(t, applyPatchSet1.Service, readPatchSet.Service)
	require.Equal(t, applyPatchSet1.EnvoyPatchSet, readPatchSet.EnvoyPatchSet)
	require.Equal(t, applyPatchSet1.Meta, readPatchSet.GetMeta())

	// delete it
	wm, err = config_entries.Delete(ApplyEnvoyPatchSet, "foo", nil)
	require.NoError(t, err)
	require.NotNil(t, wm)
	require.NotEqual(t, 0, wm.RequestTime)

	// verify deletion
	_, _, err = config_entries.Get(ApplyEnvoyPatchSet, "foo", nil)
	require.Error(t, err)
}
