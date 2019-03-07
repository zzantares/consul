package consul

import (
	"context"
	"io"
	"net"
	"sync/atomic"

	"github.com/hashicorp/go-msgpack/codec"
)

// RPCStream represents a bi-directional messaging connection where both client
// and server can send and receive arbitrary payloads at any time.
type RPCStream interface {
	// Send encodes and delivers object o to the client. It blocks until the write
	// completes successfully or an error occurs. If an error occurs, an plain
	// error object may be sent and will be interpreted by the other end as fatal
	// to the streaming call. It will be returned from Recv as an error value
	// rather than a message value.
	Send(ctx context.Context, o interface{}) error

	// Recv blocks until the next message is available or an error occurs.
	Recv(ctx context.Context, o interface{}) error

	io.Closer
}

// StreamingRPCHandler is the type of function or method that accepts streaming
// RPCs.
type StreamingRPCHandler func(stream RPCStream) error

// StreamingRPCHeader describes a streaming call, it's the first message written
// in the protocol by the client.
type StreamingRPCHeader struct {
	Method string
}

// StreamingRPCErr describes an error that has occurred. It is used rather than
// a plain error to allow discriminating between error types without parsing the
// message. We re-use http error codes just for the sake of not inventing our
// own set to do the same things.
type StreamingRPCErr struct {
	Code    int
	Message string
}

func (e *StreamingRPCErr) Error() string {
	return e.Message
}

// inMemRPCStream is used on Servers to stream between HTTP endpoints and
// Streaming RPC endpoints without the overhead of actually encoding everything
// and sending it over TCP. It is just a buffered chan. The buffering simulates
// TCP buffering so allows sending end to send some amount before blocking but
// keeps some back-pressure if the reader is slow.
type inMemRPCStream struct {
	// closed is used as an atomic flag
	closed int32
	ch     chan interface{}
}

func newInMemRPCStream() *inMemRPCStream {
	return &inMemRPCStream{
		// 256 is a number I plucked from the air
		ch: make(chan interface{}, 256),
	}
}

func (s *inMemRPCStream) Send(ctx context.Context, o interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.ch <- o:
	}
	if atomic.LoadInt32(&s.closed) == 1 {
		// Assume we only unblocked because the chan just closed.
		return io.EOF
	}
	return nil
}

func (s *inMemRPCStream) Recv(ctx context.Context, o *interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case val, ok := <-s.ch:
		if !ok {
			// Chan is closed
			return io.EOF
		}
		if err, ok := val.(error); ok {
			s.Close()
			return err
		}
		if err, ok := val.(*StreamingRPCErr); ok {
			s.Close()
			return err
		}
		*o = val
		return nil
	}
}

func (s *inMemRPCStream) Close() error {
	if old := atomic.SwapInt32(&s.closed, 1); old == 0 {
		close(s.ch)
	}
	return nil
}

type msgpackRPCStream struct {
	c   net.Conn
	enc *codec.Encoder
	dec *codec.Decoder
}

var msgpackHandle = &codec.MsgpackHandle{
	RawToString: true,
}

func newMsgpackRPCStream(conn net.Conn) *msgpackRPCStream {
	return &msgpackRPCStream{
		c: conn,
		// TODO(banks): not sure if it would be better to buffer these explicitly
		// either with our own buffers and NewEncoderBytes or by wrapping in an
		// bufio.Reader/Writer. The underlying conn is really a yamux stream which
		// may have it's own buffering...
		enc: codec.NewEncoder(conn, msgpackHandle),
		dec: codec.NewDecoder(conn, msgpackHandle),
	}
}

func (s *msgpackRPCStream) Send(ctx context.Context, o interface{}) error {
	// select {
	// case <-ctx.Done():
	// 	return ctx.Err()
	// case conn.Wri:
	// }
	// if atomic.LoadInt32(&s.closed) == 1 {
	// 	// Assume we only unblocked because the chan just closed.
	// 	return io.EOF
	// }
	return nil
}

func (s *msgpackRPCStream) Recv(ctx context.Context, o *interface{}) error {
	// select {
	// case <-ctx.Done():
	// 	return ctx.Err()
	// case val, ok := <-s.ch:
	// 	if !ok {
	// 		// Chan is closed
	// 		return io.EOF
	// 	}
	// 	if err, ok := val.(error); ok {
	// 		s.Close()
	// 		return err
	// 	}
	// 	if err, ok := val.(*StreamingRPCErr); ok {
	// 		s.Close()
	// 		return err
	// 	}
	// 	*o = val
	return nil
	//}
}

func (s *msgpackRPCStream) Close() error {
	return s.c.Close()
}
