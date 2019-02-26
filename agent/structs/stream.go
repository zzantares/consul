package structs

import (
	"encoding/binary"
	"errors"
)

type StreamTopic uint32

const (
	StreamTopicAll StreamTopic = iota
	StreamTopicKV
	StreamTopicCatalogServices

	StreamMsgTypeNop uint16 = iota
	StreamMsgTypeEvent

	StreamFrameHeaderLen int = 8
)

// StreamFrameHeader is used to frame the events sent over the wire. It's
// encoded not as msgpack but as fixed-length big-endian ints to make it fixed
// size for clients to read.
type StreamFrameHeader struct {
	Len   uint32
	Type  uint16
	Flags uint16
}

func (h *StreamFrameHeader) WriteBytes(b []byte) error {
	if len(b) < StreamFrameHeaderLen {
		return errors.New("not enough room in buffer to encode header")
	}
	binary.BigEndian.PutUint32(b[0:4], h.Len)
	binary.BigEndian.PutUint16(b[4:6], h.Type)
	binary.BigEndian.PutUint16(b[6:8], h.Flags)
	return nil
}

func (h *StreamFrameHeader) ReadBytes(b []byte) error {
	if len(b) < StreamFrameHeaderLen {
		return errors.New("not enough bytes to decode header")
	}
	h.Len = binary.BigEndian.Uint32(b[0:4])
	h.Type = binary.BigEndian.Uint16(b[4:6])
	h.Flags = binary.BigEndian.Uint16(b[6:8])
	return nil
}

type StreamEvent struct {
	Index uint64
	Topic StreamTopic
	Key   string
	Value interface{}
}
