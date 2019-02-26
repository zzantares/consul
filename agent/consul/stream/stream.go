package stream

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-msgpack/codec"
)

var ErrIndexTruncated = errors.New("index truncated from stream, restart from new snapshot")

var msgpackHandle = &codec.MsgpackHandle{
	RawToString: true,
}

// Stream stores a limited size buffer of the last events emitted by state store
// transactions. There can be at most one event per raft log Apply but may well
// be none making this a sparse set.
//
// Each event has a single topic and an identifying key within that topic.
// Topics are genearl interest areas like "KV" or "Catalog Registrations" while
// keys identify the subset of data being operated on like the Key for KV or
// service name for catalog.
//
// A reader can subscribe to a single topic and will have all messages they've
// not yet seen in that topic. They may specify a key match to only see events
// related to that specific key. When they catch up, they will only be woken to
// deliver new events if their topic and key match to save on wasted CPU cycles
// when thousands of clients are watching specific things.
//
// All topics are stored in a single buffer for now since workloads vary and
// we'd "waste" a lot of memory if all topics are sized the same but workload is
// skewed heavily towards one topic. We do store the last event index for each
// topic though so even if there a no events in the buffer for a topic, clients
// that are caught up can still know they are up-to-date and wait for new ones
// instead of re-syncing the whole state again. This limits the extent to which
// noisy topics can affect streming of quieter ones but we may still decide to
// be smarter about partitioning them later.
type Stream struct {
	size     int
	sizeMask uint64

	mu sync.Mutex

	events []Event

	// writeCount is incremented on each Publish. The index in the ring buffer
	// that will be next written to is writeCount % size (or writeCount &
	// sizeMask) which is equivalent but faster for power-of-two sizes.
	writeCount uint64

	// lastTruncatedIndex stored the raft index of the last even to be overwritten
	// in the ring buffer or 0 if we've not wrapped yet. It allows a client who
	// saw the last event that's just been overwritten to continue reading from
	// the oldest event in the buffer even if there is a gap between the events'
	// raft indices.
	lastTruncatedIndex uint64

	staged Event

	waitCh chan struct{}

	highWaterMarks map[structs.StreamTopic]uint64
}

// Event represents an event in the event stream as stored internally on
// servers. We store the actual event already msgpack encoded and framed so that
// we only do that work once and can then deliver directly to all watching
// clients.
type Event struct {
	Index   uint64
	Topic   structs.StreamTopic
	Key     string
	Payload []byte
}

// New returns a new stream capable of storing at least size events. It will
// actually choose the next highest power of two greater than size to make
// internals more efficient.
func New(size int) *Stream {
	size = nextPowerOf2(size)
	return &Stream{
		size:           size,
		sizeMask:       uint64(size - 1),
		events:         make([]Event, size),
		waitCh:         make(chan struct{}),
		highWaterMarks: make(map[structs.StreamTopic]uint64),
	}
}

// Cap returns the total number of events this Stream buffer can hold.
func (s *Stream) Cap() int {
	return s.size
}

// MinIndex returns the raft index of the oldest event in the buffer or 0 if no
// events have been written.
func (s *Stream) MinIndex() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.minIndexLocked()
}

func (s *Stream) minIndexLocked() uint64 {
	if s.writeCount == 0 {
		return 0
	}

	return s.events[s.oldestIdx()].Index
}

// MaxIndex returns the raft index of the newest event in the buffer or 0 if no
// events have been written.
func (s *Stream) MaxIndex() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.maxIndexLocked()
}

func (s *Stream) maxIndexLocked() uint64 {
	if s.writeCount == 0 {
		return 0
	}
	// find the buffer index of the last completed write (we increment count after
	// each write).
	latestIdx := (s.writeCount - 1) & s.sizeMask
	return s.events[latestIdx].Index
}

// PreparePublish sanity checks and encodes the event and stages it ready for
// publishing or returns an error. The Stream is not yet modified. If no error
// is returned, the staged event will be published by the next call to Commit.
// This allows the state store to only make events visible to clients once the
// main mem-db transaction has committed but allows catching any errors during
// encoding or duplicate publishes before mem-db commits.
func (s *Stream) PreparePublish(e structs.StreamEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.staged.Index != 0 {
		return errors.New("event already staged for commit")
	}

	// Sanity check index
	if e.Index < 1 || e.Index <= s.maxIndexLocked() {
		return errors.New("event index invalid")
	}

	// Stage event
	s.staged.Index = e.Index
	s.staged.Topic = e.Topic
	s.staged.Key = e.Key

	// Check there is a buffer with at least enough room for the header
	if len(s.staged.Payload) <= structs.StreamFrameHeaderLen {
		s.staged.Payload = make([]byte, 128)
	}

	// Create a msgpack encode that will write straight to the buffer, leaving
	// enough room for the frame header.
	payloadSlice := s.staged.Payload[structs.StreamFrameHeaderLen:]
	enc := codec.NewEncoderBytes(&payloadSlice, msgpackHandle)
	if err := enc.Encode(e); err != nil {
		return err
	}

	header := structs.StreamFrameHeader{
		Len:  uint32(len(payloadSlice)),
		Type: structs.StreamMsgTypeEvent,
	}
	if err := header.WriteBytes(s.staged.Payload[0:8]); err != nil {
		return err
	}
	// Update payload slice length to match the contents since decoder was working
	// on a different slice.
	s.staged.Payload = s.staged.Payload[0 : len(payloadSlice)+structs.StreamFrameHeaderLen]

	return nil
}

// Commit the staged event to the stream if there is one. This can never fail.
func (s *Stream) Commit() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.staged.Index == 0 {
		// Nothing staged
		return
	}

	// Mask wraps number around the buffer
	newOffset := s.writeCount & s.sizeMask

	evt := &s.events[newOffset]

	// track the index of the event we are overwriting.
	overwrittenIndex := evt.Index

	// Set the event values from staged event
	evt.Index = s.staged.Index
	evt.Topic = s.staged.Topic
	evt.Key = s.staged.Key

	// Watch closely, we actually swap the payload byte slice with the staged one
	// so that we don't let the old one go to be GCed and have to allocate a new
	// one.
	evt.Payload, s.staged.Payload = s.staged.Payload, evt.Payload

	// Update the state
	s.writeCount++

	// Update the high-water mark for this topic
	s.highWaterMarks[evt.Topic] = evt.Index

	s.lastTruncatedIndex = overwrittenIndex

	// Mark the staged event as "empty"
	s.staged.Index = 0

	// Notify watchers and create new watch chan.
	close(s.waitCh)
	s.waitCh = make(chan struct{})
}

// Abort discards the stage event if there is one. This can never fail.
func (s *Stream) Abort() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set the index to 0 to signal that it's a non-event. We don't use a pointer
	// because we can save on allocations if we re-use the Event struct and
	// payload slice.
	s.staged.Index = 0
}

// Publish pushes a new event into the buffer. An error is returned if the
// payload can't be encoded or if there was already an event published for this
// raft index or a higher one.
func (s *Stream) Publish(e structs.StreamEvent) error {
	err := s.PreparePublish(e)
	if err != nil {
		return err
	}
	s.Commit()
	return nil
}

func (s *Stream) Next(index uint64, topic structs.StreamTopic, key string) (uint64, []byte, error) {
	return s.next(context.Background(), false, index, topic, key)
}

func (s *Stream) NextBlock(ctx context.Context, index uint64, topic structs.StreamTopic, key string) (uint64, []byte, error) {
	return s.next(ctx, true, index, topic, key)
}

func (s *Stream) next(ctx context.Context, blocking bool, index uint64, topic structs.StreamTopic, key string) (uint64, []byte, error) {
	// TODO(banks) benchmark concurrency. Blocking or starving the writer which is
	// sequntial with raft apply is not ideal but it way be overkill to try any
	// make this all lock-free at first pass. RWMutex _might_ help with many
	// readers but it would be just as likely if not more likely to starve the
	// writer thread. Atomic update on the ring indexes are possible but a lot
	// more subtle and tend to need busy loops to wait which would get expensive
	// fast.
	originalIndex := index

	nextIdx := uint64(s.size)

	s.mu.Lock()
	for {
		maxIndex := s.maxIndexLocked()
		if topic != structs.StreamTopicAll {
			maxIndex = s.highWaterMarks[topic]
		}
		if maxIndex <= index {
			// We reached the head or there has been no event in this topic ever so we
			// need to wait for one. Copy the current waitCh while we still hold the
			// lock as it gets written under lock during publish.
			waitCh := s.waitCh
			s.mu.Unlock()
			if !blocking {
				return originalIndex, nil, nil
			}
			select {
			case <-ctx.Done():
				return originalIndex, nil, ctx.Err()
			case <-waitCh:
				// continue loop as there is a
			}
			// All waiters will have just unblocked together and now will contend over
			// this lock to make it around the loop again. Short of complicated
			// lock-free stuff or RWMutex that might not help or cause Publish to be
			// starved more often, we could at least improve things by making separate
			// waitCh by topic and key so we only unblock waiters who actually have
			// useful work to do.
			s.mu.Lock()
		}

		// Check that the index we last saw is still in the buffer. We compare
		// against the index of the last even we dropped so that the oldest event
		// left is still useful. If events have been dropped we will need to refetch
		// a full snapshot. index == lastTruncatedIndex implies that we did see that
		// last event therefore the oldest one still here is the next one and we can
		// continue. < is used rather than == to allow clients who are getting
		// indexes out of band (e.g. from other result sets where the index returned
		// is not the exact one at which an event was emitted) provided it's more
		// recent than the last thing we dropped.
		if index < s.lastTruncatedIndex {
			s.mu.Unlock()
			return index, nil, ErrIndexTruncated
		}

		// Now we know there are more events for our topic to check out, we need to
		// find the next event we've not seen yet. If we've been around this loop
		// before then it's just the next one in the buffer so we don't need to
		// waste time searching. If this is the first iteration, then we need to
		// locate the event immediately after the index we passed which we can do
		// with binary search. nextIdx will always be less than s.size once we've
		// been round the loop once.
		if nextIdx == uint64(s.size) {
			nextIdx = s.nextBufferIdxAfter(index)
			if nextIdx == uint64(s.size) {
				// Not found! This shouldn't be possible given the checks above but
				// rather than blow up in case something went strange, just start at the
				// oldest one and iterate the whole buffer. This is always correct but
				// possibly slow.
				nextIdx = s.oldestIdx()
			}
		}

		// See if the next index is relevant to us
		nextEvt := &s.events[nextIdx]

		// Note that we need to check
		if (topic == structs.StreamTopicAll || nextEvt.Topic == topic) &&
			(key == "" || key == nextEvt.Key) {
			s.mu.Unlock()
			return nextEvt.Index, nextEvt.Payload, nil
		}

		// That event was not for us, update our index and loop around.
		nextIdx = (nextIdx + 1) & s.sizeMask
		// Update the index we've "got to" so that we can check against the maxIndex
		// on next iteration.
		index = nextEvt.Index
	}
}

// oldestIdx returns the _buffer_ index (i.e. wrapped to s.size) of the oldest
// event currently in the buffer. It returns s.size if there are no events in
// the buffer.
func (s *Stream) oldestIdx() uint64 {
	if s.writeCount == 0 {
		return uint64(s.size)
	}
	if s.writeCount < uint64(s.size) {
		return 0
	}
	return s.writeCount & s.sizeMask
}

// nextBufferIdxAfter finds the buffer index for the next event after the given
// raft index. We can do that efficiently with binary search although it's a bit
// confusing since the "sorted" range of the buffer wraps around.
func (s *Stream) nextBufferIdxAfter(index uint64) uint64 {
	oldestIdx := s.oldestIdx()
	if oldestIdx == uint64(s.size) {
		// return the size of the array as "no such index"
		return oldestIdx
	}
	size := s.size
	// If we've not wrapped yet, only consider the offsets with value Events in.
	if s.writeCount < uint64(s.size) {
		size = int(s.writeCount)
	}
	nextOffset := sort.Search(size, func(i int) bool {
		return s.events[(oldestIdx+uint64(i))&s.sizeMask].Index > index
	})
	// nextOffset is now the offset from the minimum event which may need to wrap
	// around the buffer.
	return (uint64(nextOffset) + oldestIdx) & s.sizeMask
}

// block must be called when s.mu is NOT held by the caller.
func (s *Stream) block(ctx context.Context, index uint64, offset, topic structs.StreamTopic, key string) error {
	// TODO
	s.mu.Lock()

	// Double check we are still at the most recent index since acquireing the
	// lock again.
	maxIndex := s.highWaterMarks[topic]
	if maxIndex > 0 && maxIndex > index {
		// There is already a new event for this topic
	}

	return nil
}

func nextPowerOf2(size int) int {
	if size < 2 || size > (1<<31) {
		// Just pick a sane value if the input is wild. Shouldn't happen as we'll
		// validate elsewhere.
		return 128
	}

	// Need a uint32 for this to work.
	sz := uint32(size)

	// Decrement in case size is already a power of 2 - we don't want to round up
	// to the next one.
	sz--

	// Fill in all the bits on the right of the most significant bit. To make this
	// more educational and less magic we'll annotate the shifts with the effect
	// they have on an example where size is 16777473:
	// sz 	    = 00000001 00000000 00000001 00000000
	// sz >> 1  = 00000000 10000000 00000000 10000000
	sz |= sz >> 1
	// sz 	    = 00000001 10000000 00000001 10000001
	// sz >> 2  = 00000000 01100000 00000000 01100000
	sz |= sz >> 2
	// sz 	    = 00000001 11100000 00000001 11100001
	// sz >> 4  = 00000000 00011110 00000000 00011110
	sz |= sz >> 4
	// sz 	    = 00000001 11111110 00000001 11111111
	// sz >> 8  = 00000000 00000001 11111110 00000001
	sz |= sz >> 8
	// sz       = 00000001 11111111 11111111 11111111
	// Note: in the example we now have all 1s to right of most significant bit
	// but if there were no bits set in the lower 2 bytes to start with.
	// sz >> 16 = 00000000 00000000 00000001 11111111
	sz |= sz >> 16
	// sz       = 00000001 11111111 11111111 11111111
	// This is 1 less than the next highest power of 2!
	sz++
	return int(sz)
}
