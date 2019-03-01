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

	// Default to making the store large enough to fit all the events just so we
	// don't loose any data on restore. In practice the state store will never
	// have a nil stream, this is just to handle the case where stream is nil
	// mostly in tests etc.
	size := len(snap.Events)

	// Use the state store's existing stream size if there is one. This allows
	// reconfiguration of that stream size to take effect on a restart since the
	// state store's stream is initialised to a size from config on startup.
	if r.store.stream != nil {
		size = r.store.stream.Cap()
	}
	r.store.stream, err = stream.NewFromSnapshot(size, snap)
	return err
}

// emitEvent publishes an event to the store's stream if and only if txn
// commits. Event's index must be set to the current raft index and it is an
// error to call this more than once during the processing of a specific raft
// command.
func (s *Store) emitEvent(txn *memdb.Txn, e structs.StreamEvent) error {
	if s.stream == nil {
		return nil
	}
	err := s.stream.PreparePublish(e)
	if err != nil {
		return err
	}
	txn.Defer(s.stream.Commit)
	return nil
}
