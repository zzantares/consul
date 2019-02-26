package stream

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/stretchr/testify/require"
)

func TestStreamSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantCap int
	}{
		{
			name:    "power of two",
			size:    1024,
			wantCap: 1024,
		},
		{
			name:    "small, non-power",
			size:    3,
			wantCap: 4,
		},
		{
			name:    "medium, non-power",
			size:    100000000,
			wantCap: 1 << 27,
		},
		{
			name:    "huge, non-power",
			size:    4000000000,
			wantCap: 128,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.size > 1000000 {
				// Don't try to allocate massive arrays, just test the power-of-2 helper
				// directly.
				require.Equal(t, tt.wantCap, nextPowerOf2(tt.size))
			} else {
				s := New(tt.size)
				require.Equal(t, tt.wantCap, s.Cap())
			}
		})
	}
}

var eventIndex uint64

func makeEvent(key string) structs.StreamEvent {
	return makeEventInTopic(key, structs.StreamTopicKV)
}

func makeEventInTopic(key string, topic structs.StreamTopic) structs.StreamEvent {
	return structs.StreamEvent{
		Index: atomic.AddUint64(&eventIndex, 100), // Ensured indexes are sparse
		Topic: topic,
		Key:   key,
		Value: key,
	}
}

func decodeStreamEventFrame(t *testing.T, b []byte) structs.StreamEvent {
	t.Helper()
	// Decode the frame header
	var header structs.StreamFrameHeader
	require.NoError(t, header.ReadBytes(b))
	require.Equal(t, uint32(len(b)-structs.StreamFrameHeaderLen), header.Len)
	require.Equal(t, header.Flags, uint16(0))
	require.Equal(t, header.Type, structs.StreamMsgTypeEvent)

	// Decode the msgpack value
	var e structs.StreamEvent
	dec := codec.NewDecoderBytes(b[structs.StreamFrameHeaderLen:], msgpackHandle)
	require.NoError(t, dec.Decode(&e))
	return e
}

func assertEventPayload(t *testing.T, b []byte, e structs.StreamEvent) {
	t.Helper()
	got := decodeStreamEventFrame(t, b)
	require.Equal(t, e, got)
}

func TestStreamPublishNext(t *testing.T) {
	s := New(4)

	require := require.New(t)

	// Next on an empty stream should return nil (and not block)
	idx, got, err := s.Next(0, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Nil(got)
	require.Equal(uint64(0), idx)

	// Publish a message
	e1 := makeEvent("one")
	require.NoError(s.Publish(e1))
	require.Equal(e1.Index, s.MinIndex())

	// Next should return it if we use an older index.
	idx, got, err = s.Next(s.MinIndex()-1, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e1.Index, idx)
	assertEventPayload(t, got, e1)

	// Since we just initialised the stream, and index including 0 up to the first
	// even should be sufficient as we know there were no events earlier.
	idx, got, err = s.Next(0, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e1.Index, idx)
	assertEventPayload(t, got, e1)

	// Now fill up the buffer with events until it overflows
	e2 := makeEvent("two")
	require.NoError(s.Publish(e2))
	require.Equal(e1.Index, s.MinIndex())
	e3 := makeEvent("three")
	require.NoError(s.Publish(e3))
	require.Equal(e1.Index, s.MinIndex())
	e4 := makeEvent("four")
	require.NoError(s.Publish(e4))
	require.Equal(e1.Index, s.MinIndex())
	e5 := makeEvent("five")
	require.NoError(s.Publish(e5))
	// Notice we should now have overwritten e1 and e2 should be the oldest event
	require.Equal(e2.Index, s.MinIndex())

	// Writing an older event should fail
	require.Error(s.Publish(e3))
	// Writing an event with the same index as the current max should too
	require.Error(s.Publish(e5))

	// Now we have overflowed, the Stream should know if we've missed events so
	// starting from 0 should fail since we'd have missed e1.
	idx, got, err = s.Next(0, structs.StreamTopicKV, "")
	require.Error(err)
	require.Equal(ErrIndexTruncated, err)
	require.Nil(got)
	require.Equal(uint64(0), idx)

	// Consumer should be able to read through all the stuff from e1 onwards even
	// though e1 is not in the buffer any more.
	idx, got, err = s.Next(e1.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e2.Index, idx)
	assertEventPayload(t, got, e2)

	idx, got, err = s.Next(e2.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e3.Index, idx)
	assertEventPayload(t, got, e3)

	idx, got, err = s.Next(e3.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e4.Index, idx)
	assertEventPayload(t, got, e4)

	idx, got, err = s.Next(e4.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e5.Index, idx)
	assertEventPayload(t, got, e5)

	// Then when caught up should return nil and the same index
	idx, got, err = s.Next(e5.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e5.Index, idx)
	require.Nil(got)

	// And if we try to resume for an old event (e1)

	/////////////
	// Topic filtering

	// An event published to a different topic should not be seen by Next
	e6 := makeEventInTopic("six", structs.StreamTopicCatalogServices)
	require.NoError(s.Publish(e6))
	require.Equal(e3.Index, s.MinIndex())

	// Next on the other topic should still return nil (caught up)
	idx, got, err = s.Next(e5.Index, structs.StreamTopicKV, "")
	require.NoError(err)
	require.Equal(e5.Index, idx)
	require.Nil(got)

	// Next on the new topic should return the new event
	idx, got, err = s.Next(e5.Index, structs.StreamTopicCatalogServices, "")
	require.NoError(err)
	require.Equal(e6.Index, idx)
	assertEventPayload(t, got, e6)

	/////////////
	// All Topic

	// Should see everything (i.e. both e5 and e6 if starting from e4 index)
	idx, got, err = s.Next(e4.Index, structs.StreamTopicAll, "")
	require.NoError(err)
	assertEventPayload(t, got, e5)
	require.Equal(e5.Index, idx)

	idx, got, err = s.Next(e5.Index, structs.StreamTopicAll, "")
	require.NoError(err)
	assertEventPayload(t, got, e6)
	require.Equal(e6.Index, idx)

	// Then when caught up should return nil and the same index
	idx, got, err = s.Next(e6.Index, structs.StreamTopicAll, "")
	require.NoError(err)
	require.Nil(got)
	require.Equal(e6.Index, idx)

	/////////////
	// Key filtering

	// We currently have e3-e6 inclusive in the buffer all with different keys.
	// Next with a key filter on any one should only get one result...
	idx, got, err = s.Next(e3.Index, structs.StreamTopicKV, "four")
	require.NoError(err)
	assertEventPayload(t, got, e4)
	require.Equal(e4.Index, idx)
	// Then should not get anything else (caught up)
	idx, got, err = s.Next(e4.Index, structs.StreamTopicKV, "four")
	require.NoError(err)
	require.Nil(got)
	require.Equal(e4.Index, idx)

	// But a Next filtering on "five" with otherwise identical argumesnt should
	// see e5.
	idx, got, err = s.Next(e4.Index, structs.StreamTopicKV, "five")
	require.NoError(err)
	assertEventPayload(t, got, e5)
	require.Equal(e5.Index, idx)
	// Then should not get anything else (caught up)
	idx, got, err = s.Next(e5.Index, structs.StreamTopicKV, "five")
	require.NoError(err)
	require.Nil(got)
	require.Equal(e5.Index, idx)
}

type retVals struct {
	Index   uint64
	Payload []byte
	Err     error
}

// nolint - context is a passthrough param and would be more confusing to put
// first in this helper.
func runBlockingNext(t *testing.T, s *Stream, ctx context.Context, index uint64, topic structs.StreamTopic, key string) chan retVals {
	t.Helper()
	ch := make(chan retVals, 1)
	go func() {
		var v retVals
		v.Index, v.Payload, v.Err = s.NextBlock(ctx, index, topic, key)
		ch <- v
	}()
	return ch
}

func assertBlocked(t *testing.T, ch chan retVals) {
	t.Helper()
	select {
	case <-time.After(50 * time.Millisecond):
		// Seems to have blocked OK!
	case v := <-ch:
		t.Fatalf("should have blocked, got %v", v)
	}
}

func assertDeliveredSoon(t *testing.T, ch chan retVals, e structs.StreamEvent) {
	t.Helper()
	select {
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timed out waiting for delivery")
	case v := <-ch:
		require.NoError(t, v.Err)
		assertEventPayload(t, v.Payload, e)
		require.Equal(t, e.Index, v.Index)
	}
}

func TestStreamNextBlocking(t *testing.T) {
	s := New(4)

	require := require.New(t)

	// Blocking client should block with nothing in the stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create event here so we can wait on the event immediately before it.
	e1 := makeEvent("one")

	ch := runBlockingNext(t, s, ctx, e1.Index-1, structs.StreamTopicKV, "")
	assertBlocked(t, ch)

	// Publishing a message should unblock the blocking listener and deliver it.
	require.NoError(s.Publish(e1))
	assertDeliveredSoon(t, ch, e1)

	// Wait again on the e1 index should block
	ch = runBlockingNext(t, s, ctx, e1.Index, structs.StreamTopicKV, "")
	assertBlocked(t, ch)

	// Publishing a message for a different topic should NOT unblock
	e2 := makeEventInTopic("two", structs.StreamTopicCatalogServices)
	require.NoError(s.Publish(e2))

	assertBlocked(t, ch)

	// But another event in the same topic should unblock and return that one
	e3 := makeEvent("three")
	require.NoError(s.Publish(e3))
	assertDeliveredSoon(t, ch, e3)

	// Run a blocker with a key filter for "one" and index back at the start it
	// should immediately return e1.
	ch = runBlockingNext(t, s, ctx, e1.Index-1, structs.StreamTopicKV, "one")
	assertDeliveredSoon(t, ch, e1)

	// But then block (and not return e3 despite same topic)
	ch = runBlockingNext(t, s, ctx, e1.Index, structs.StreamTopicKV, "one")
	assertBlocked(t, ch)

	// Eventually a new event with same key and topic should unblock it
	e4 := makeEvent("one")
	require.NoError(s.Publish(e4))
	assertDeliveredSoon(t, ch, e4)

	// Blocked requests should be cancellable/timeoutable
	ch = runBlockingNext(t, s, ctx, e4.Index, structs.StreamTopicKV, "")
	assertBlocked(t, ch)

	cancel()

	select {
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timed out waiting for cancellation")
	case v := <-ch:
		require.Error(v.Err)
		require.Equal(ctx.Err(), v.Err)
		require.Nil(v.Payload)
		require.Equal(e4.Index, v.Index)
	}
}

func TestStreamPrepareCommit(t *testing.T) {
	s := New(4)

	require := require.New(t)

	// Blocking client should block with nothing in the stream
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create event here so we can wait on the event immediately before it.
	e1 := makeEvent("one")
	e2 := makeEvent("two")
	e3 := makeEvent("three")

	ch := runBlockingNext(t, s, ctx, e1.Index-1, structs.StreamTopicKV, "")
	assertBlocked(t, ch)

	// Preparing an event should NOT unblock the blocking listener.
	require.NoError(s.PreparePublish(e1))
	assertBlocked(t, ch)

	// Preparing a second event before commit should fail
	err := s.PreparePublish(e2)
	require.Error(err)
	require.Contains(err.Error(), "event already staged for commit")
	assertBlocked(t, ch)

	// Aborting the staged event should allow a new event to be staged.
	s.Abort()

	require.NoError(s.PreparePublish(e3))
	// But should not be delivered yet
	assertBlocked(t, ch)

	// Commit should publish event and allow it to be delivered
	s.Commit()
	assertDeliveredSoon(t, ch, e3)

	// Committing with nothing staged should be a no-op and not deliver anything
	min, max := s.MinIndex(), s.MaxIndex()
	s.Commit()
	assertBlocked(t, ch)
	require.Equal(min, s.MinIndex())
	require.Equal(max, s.MaxIndex())

	// Aborting with nothing staged should be a no-op and not deliver anything
	s.Abort()
	assertBlocked(t, ch)
	require.Equal(min, s.MinIndex())
	require.Equal(max, s.MaxIndex())
}

func TestStreamConcurrentReaders(t *testing.T) {
	s := New(1024)

	// Setup a bunch of readers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventKeys := []string{"foo", "bar", "baz", "qux"}

	count := uint32(0)

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	nWatchers := 1000
	nMessages := 200

	for i := 0; i < nWatchers; i++ {
		wg.Add(1)
		go func(key string) {
			idx := uint64(99) // hack we know first event will have index of 100
			for {
				nextIdx, _, err := s.NextBlock(ctx, idx, structs.StreamTopicKV, key)
				if err != nil {
					if err != ctx.Err() {
						errs <- err
					}
					wg.Done()
					return
				}

				idx = nextIdx
				atomic.AddUint32(&count, 1)
			}
		}(eventKeys[i%4])
	}

	// Now publish lots of messages
	for i := 0; i < nMessages; i++ {
		e := makeEvent(eventKeys[i%4])
		require.NoError(t, s.Publish(e))
		// We have to pace ourselves otherwise we push the events faster than
		// consumer can read them and then they fall behind
		time.Sleep(10 * time.Microsecond)
	}

	// Wait a bit longer for all the deliveries to be made.
	time.Sleep(50 * time.Millisecond)

	cancel()

	wg.Wait()

	// Should be no errors in clients
	select {
	case err := <-errs:
		require.NoError(t, err)
	default:
	}

	// Each of our 20000 messsages to 4 topics should have been delivered to 1/4
	// of the 100 subscribers.
	require.Equal(t, uint32(nMessages*nWatchers/len(eventKeys)), count)
}
