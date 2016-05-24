package client

import (
	"errors"
	"io"
	"sync"

	"github.com/lostmyname/golumn/metrics"

	"golang.org/x/net/context"
)

// Implements the streamer interface
type rpcStream struct {
	sync.RWMutex
	seq     uint64
	closed  chan bool
	err     error
	request Request
	codec   clientCodec
	context context.Context
}

func (r *rpcStream) isClosed() bool {
	select {
	case <-r.closed:
		return true
	default:
		return false
	}
}

func (r *rpcStream) Context() context.Context {
	return r.context
}

func (r *rpcStream) Request() Request {
	return r.request
}

func (r *rpcStream) Send(msg interface{}) error {
	r.Lock()
	defer r.Unlock()

	if r.isClosed() {
		r.err = errShutdown
		return errShutdown
	}

	seq := r.seq
	r.seq++

	req := request{
		Service:       r.request.Service(),
		Seq:           seq,
		ServiceMethod: r.request.Method(),
	}

	if err := r.codec.WriteRequest(&req, msg); err != nil {
		r.err = err
		return err
	}
	return nil
}

func (r *rpcStream) Recv(msg interface{}) error {
	r.Lock()
	defer r.Unlock()

	if r.isClosed() {
		r.err = errShutdown
		return errShutdown
	}

	rrhs := metrics.NewTiming()
	var resp response
	if err := r.codec.ReadResponseHeader(&resp); err != nil {
		if err == io.EOF && !r.isClosed() {
			r.err = io.ErrUnexpectedEOF
			return io.ErrUnexpectedEOF
		}
		r.err = err
		return err
	}
	rrhs.Send("micro.client.stream.recv.readresponse.time")

	switch {
	case len(resp.Error) > 0:
		rehs := metrics.NewTiming()
		// We've got an error response. Give this to the request;
		// any subsequent requests will get the ReadResponseBody
		// error if there is one.
		if resp.Error != lastStreamResponseError {
			r.err = serverError(resp.Error)
		} else {
			r.err = io.EOF
		}
		rehrrs := metrics.NewTiming()
		if err := r.codec.ReadResponseBody(nil); err != nil {
			r.err = errors.New("reading error payload: " + err.Error())
		}
		rehrrs.Send("micro.client.stream.recv.error.readresponsebody.time")
		rehs.Send("micro.client.stream.recv.error.response.time")
	default:
		rnehs := metrics.NewTiming()
		if err := r.codec.ReadResponseBody(msg); err != nil {
			r.err = errors.New("reading body " + err.Error())
		}
		rnehs.Send("micro.client.stream.recv.noerror.resonse.time")
	}

	return r.err
}

func (r *rpcStream) Error() error {
	r.RLock()
	defer r.RUnlock()
	return r.err
}

func (r *rpcStream) Close() error {
	select {
	case <-r.closed:
		return nil
	default:
		close(r.closed)
		return r.codec.Close()
	}
}
