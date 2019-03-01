package state

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/consul/stream"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/require"
)

func TestStateStore_StreamEmit(t *testing.T) {
	require := require.New(t)

	s, err := NewStateStoreWithStream(nil, stream.New(4))
	require.NoError(err)

	tx := emitEvent(t, s, 1234)

	// Transaction has not committed yet so we should _not_ see the publish yet.
	idx, got, err := s.stream.Next(0, structs.StreamTopicKV, "test")
	require.NoError(err)
	require.Nil(got)
	require.Equal(uint64(0), idx)

	// Transaction commits
	tx.Commit()

	// We should see the published event soon.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		idx, got, err := s.stream.NextBlock(ctx, 0, structs.StreamTopicKV, "test")
		require.NoError(err)
		require.NotNil(got) // Assume the payload is correct as that is tested elsewhere
		require.Equal(uint64(1234), idx)
	}

	// Now try a transaction that aborts
	tx = emitEvent(t, s, 2345)

	// Transaction has not committed yet so we should _not_ see the publish yet.
	idx, got, err = s.stream.Next(1234, structs.StreamTopicKV, "test")
	require.NoError(err)
	require.Nil(got)
	require.Equal(uint64(1234), idx)

	// Transaction ABORTS
	tx.Abort()

	// We should NOT see the published event soon.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		idx, got, err := s.stream.NextBlock(ctx, 1234, structs.StreamTopicKV, "test")
		require.Error(err)
		require.Equal(context.DeadlineExceeded, err)
		require.Nil(got)
		require.Equal(uint64(1234), idx)
	}

	// Following an abort, the staged event will need to be cleaned up. ApplyDone
	// is now called by the FSM after every Apply handler returns. So we need to
	// make sure we can emit a new event again once it is. Note that if we don't
	// call this, then emitEvent below will error and fail the test.
	s.ApplyDone()

	tx = emitEvent(t, s, 2345)
	tx.Abort()
}

func emitEvent(t *testing.T, s *Store, index uint64) *memdb.Txn {
	t.Helper()
	tx := s.db.Txn(true)

	// Just need to update the index on one of the "table" indexes which are used
	// when working out the snapshot index. We'll pick the index of the index
	// table itself for funsies.
	require.NoError(t, tx.Insert("index", &IndexEntry{"index", index}))
	require.NoError(t, s.emitEvent(tx, structs.StreamEvent{
		Index: index,
		Topic: structs.StreamTopicKV,
		Key:   "test",
		Value: "foo",
	}))

	return tx
}

func TestStateStore_SnapshotCapturesStream(t *testing.T) {
	require := require.New(t)

	s, err := NewStateStoreWithStream(nil, stream.New(4))
	require.NoError(err)

	// Populate with some events (event Raft indexes will be 1-10 inclusive)
	for i := 0; i < 10; i++ {
		emitEvent(t, s, uint64(i+1)).Commit()
	}

	// Snapshot of the state store should populate the stream snapshot at the
	// right index.
	snap := s.Snapshot()

	ss := snap.Stream()
	require.NotNil(ss)
	require.Len(ss.Events, 4)
	require.Equal(uint64(10), ss.Events[3].Index)
	require.Equal(uint64(10), ss.HighWaterMarks[structs.StreamTopicKV])
}
