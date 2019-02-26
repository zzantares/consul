package stream

import (
	"errors"

	"github.com/hashicorp/consul/agent/structs"
)

// Snapshot is a representation of a Stream's state suitable for serialising
// into a snapshot via msgpack.
type Snapshot struct {
	Events             []Event
	HighWaterMarks     map[structs.StreamTopic]uint64
	LastTruncatedIndex uint64
}

// Snapshot returns a Snapshot of all events publish at or before the snapIndex
// with a copy of the events and metadata suitable for serialising to disk.
func (s *Stream) Snapshot(snapIndex uint64) *Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap := &Snapshot{
		HighWaterMarks: make(map[structs.StreamTopic]uint64),
	}

	for topic, index := range s.highWaterMarks {
		// Only add high water marks that are before the oldest event in the current
		// stream - the rest are rebuilt dynamically from the events in the loop
		// below to make sure they reflect reality at the point in time
		// snapshotIndex was made.
		if index < s.minIndexLocked() {
			snap.HighWaterMarks[topic] = index
		}
	}

	snap.LastTruncatedIndex = s.lastTruncatedIndex

	// Need to iterate from the oldest index to the newest
	idx := s.oldestIdx()
	size := s.size
	// If we've not wrapped yet, only consider the offsets with value Events in.
	if s.writeCount < uint64(s.size) {
		size = int(s.writeCount)
	}

	for len(snap.Events) < size {
		e := s.events[idx]
		if e.Index > snapIndex {
			// Don't add events that happened after the snapshot index
			break
		}
		snap.Events = append(snap.Events, e)
		snap.HighWaterMarks[e.Topic] = e.Index
		idx = (idx + 1) & s.sizeMask
	}
	return snap
}

// NewFromSnapshot creates a new Stream of given size and plays back the events
// from a saved Snapshot. Note that the size is not linked to the old size
// allowing the size of the buffer to be changes between snapshot and restore.
// If the buffer now is smaller than when it was when the snapshot was made, the
// oldest events will be dropped.
func NewFromSnapshot(size int, snap *Snapshot) (*Stream, error) {
	if snap == nil {
		return nil, errors.New("snapshot is nil")
	}
	s := New(size)

	s.lastTruncatedIndex = snap.LastTruncatedIndex

	events := snap.Events
	snapSize := len(events)
	if size < snapSize {
		// Only copy the oldest events
		events = events[snapSize-size:]

		// Fix up last truncated index too
		s.lastTruncatedIndex = snap.Events[snapSize-size-1].Index
	}

	// Copy the events direct
	for idx, e := range events {
		s.events[idx] = e
		s.writeCount++
	}

	if snap.HighWaterMarks != nil {
		s.highWaterMarks = snap.HighWaterMarks
	}

	return s, nil
}
