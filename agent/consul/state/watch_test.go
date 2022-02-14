package state

import (
	"testing"

	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/consul/agent/structs"
)

func TestWatches(t *testing.T) {
	store := NewStateStore(nil)

	err := store.EnsureRegistration(1, &structs.RegisterRequest{
		Node:    "node1",
		Service: &structs.NodeService{Service: "flavor", ID: "flavor"},
		Checks: []*structs.HealthCheck{
			{CheckID: "01", Node: "node1", ServiceName: "flavor", ServiceID: "flavor", Timeout: "10s"},
		},
	})
	require.NoError(t, err)

	err = store.EnsureRegistration(2, &structs.RegisterRequest{
		Node:    "node1",
		Service: &structs.NodeService{Service: "float", ID: "float"},
		Checks: []*structs.HealthCheck{
			{CheckID: "02", Node: "node1", ServiceName: "float", ServiceID: "float", Timeout: "10s"},
		},
	})
	require.NoError(t, err)

	t.Run("fire when new service is created", func(t *testing.T) {
		ws := memdb.NewWatchSet()
		_, _, err = store.ServiceChecks(ws, "float", nil)
		require.NoError(t, err)
		require.True(t, !watchFired(ws))

		err = store.EnsureRegistration(4, &structs.RegisterRequest{
			Node:    "node1",
			Service: &structs.NodeService{Service: "fire", ID: "fire"},
			Checks: []*structs.HealthCheck{
				{CheckID: "02", Node: "node1", ServiceName: "fire", ServiceID: "fire", Timeout: "10s"},
			},
		})
		require.NoError(t, err)

		// This assertion fails
		require.True(t, !watchFired(ws))
	})

	t.Run("fire when a different service is removed", func(t *testing.T) {
		ws := memdb.NewWatchSet()
		_, _, err = store.ServiceChecks(ws, "float", nil)
		require.NoError(t, err)
		require.True(t, !watchFired(ws))

		err = store.DeleteService(5, "node1", "flavor", nil)
		require.NoError(t, err)

		// This assertion fails
		require.True(t, !watchFired(ws))
	})
}
