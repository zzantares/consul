package state

import (
	"github.com/hashicorp/consul/agent/consul/stream"
	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-memdb"
)

// Stream returns a stream.Snapshot captured at the snapshot index.
func (s *Snapshot) Stream() *stream.Snapshot {
	return s.streamSnap
}

// StreamRestore restores the store's stream from the given snapshot.
func (r *Restore) StreamRestore(snap *stream.Snapshot) error {
	var err error
	r.store.stream, err = stream.NewFromSnapshot(r.store.stream.Cap(), snap)
	return err
}

// emitEvent publishes an event to the store's stream if and only if txn
// commits. Event's index must be set to the current raft index and it is an
// error to call this more than once during the processing of a specific raft
// command.
func (s *Store) emitEvent(txn memdb.Txn, e structs.StreamEvent) error {
	err := s.stream.PreparePublish(e)
	if err != nil {
		return err
	}
	txn.Defer(s.stream.Commit)
	return nil
}
