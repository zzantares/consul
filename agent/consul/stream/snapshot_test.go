package stream

import (
	"testing"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/stretchr/testify/require"
)

func TestStreamSnapshot(t *testing.T) {

	events := []structs.StreamEvent{}
	keys := []string{"foo", "bar", "baz"}
	for i := 0; i < 10; i++ {
		events = append(events, makeEvent(keys[i%len(keys)]))
	}

	tests := []struct {
		name          string
		size          int
		snapIndex     uint64
		events        []structs.StreamEvent
		restoreSize   int
		wantEvents    []structs.StreamEvent
		wantLastTrunc uint64
		wantHWM       map[structs.StreamTopic]uint64
	}{
		{
			name:          "empty",
			size:          4,
			snapIndex:     1 << 50, // Big number to get all events
			events:        []structs.StreamEvent{},
			restoreSize:   4,
			wantEvents:    []structs.StreamEvent{},
			wantLastTrunc: 0,
			wantHWM:       map[structs.StreamTopic]uint64{},
		},
		{
			name:          "partial",
			size:          4,
			snapIndex:     1 << 50, // Big number to get all events
			events:        events[0:3],
			restoreSize:   4,
			wantEvents:    events[0:3],
			wantLastTrunc: 0,
			wantHWM: map[structs.StreamTopic]uint64{
				structs.StreamTopicKV: events[2].Index,
			},
		},
		{
			name:          "wrapped",
			size:          4,
			snapIndex:     1 << 50, // Big number to get all events
			events:        events[0:10],
			restoreSize:   4,
			wantEvents:    events[6:10],
			wantLastTrunc: events[5].Index,
			wantHWM: map[structs.StreamTopic]uint64{
				structs.StreamTopicKV: events[9].Index,
			},
		},
		{
			name:          "resized larger",
			size:          4,
			snapIndex:     1 << 50, // Big number to get all events
			events:        events[0:10],
			restoreSize:   6,
			wantEvents:    events[6:10],
			wantLastTrunc: events[5].Index,
			wantHWM: map[structs.StreamTopic]uint64{
				structs.StreamTopicKV: events[9].Index,
			},
		},
		{
			name:          "resized smaller",
			size:          4,
			snapIndex:     1 << 50, // Big number to get all events
			events:        events[0:10],
			restoreSize:   3,
			wantEvents:    events[7:10],
			wantLastTrunc: events[6].Index,
			wantHWM: map[structs.StreamTopic]uint64{
				structs.StreamTopicKV: events[9].Index,
			},
		},
		{
			// This test simulates a state store snapshot being taken where the mem-db
			// transaction points toa version of the state store at index while a
			// concurrent FSM apply had published a new event at a higher index. We
			// need to be consistent with the state store as of the mem-db index so
			// the snapshot _should not_ include the more recent event even if that
			// means we aren't quite "full". This makes the stream snapshot not 100%
			// consistent with a point-in-time snapshot since we have already dropped
			// some event that was in the stream at the time of the memdb snapshot,
			// but it's not inconsistent, only that we might have slightly fewer
			// events in the buffer than the maximum limit.
			name: "concurrently published events",
			size: 4,
			// Snapshot is taken when memdb is between event[8]'s index and
			// event[9]'s. Event[9] is now published but shouldn't be in the
			// snapshot.
			snapIndex:     events[9].Index - 1,
			events:        events[0:10],
			restoreSize:   4,
			wantEvents:    events[6:9], // Should NOT restore event[9]
			wantLastTrunc: events[5].Index,
			wantHWM: map[structs.StreamTopic]uint64{
				// HWM should be the previous event's index
				structs.StreamTopicKV: events[8].Index,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.size)

			require := require.New(t)

			for _, e := range tt.events {
				require.NoError(s.Publish(e))
			}

			// Take snapshot
			snap := s.Snapshot(tt.snapIndex)

			// Restore into new stream
			rs, err := NewFromSnapshot(tt.restoreSize, snap)
			require.NoError(err)

			// Fetch all. We do this instead of just taking another snapshot since
			// that is also code under test here so it's more robust to directly
			// access the state.
			idx := rs.MinIndex() - 1
			got := make([]structs.StreamEvent, 0, len(tt.wantEvents))
			for {
				curIdx, payload, err := rs.Next(idx, structs.StreamTopicAll, "")
				require.NoError(err)
				if payload == nil {
					break
				}
				got = append(got, decodeStreamEventFrame(t, payload))
				idx = curIdx
			}

			require.Equal(tt.wantEvents, got)
			require.Equal(tt.wantLastTrunc, rs.lastTruncatedIndex)
			require.Equal(tt.wantHWM, rs.highWaterMarks)
		})
	}
}
