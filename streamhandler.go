package simplelog

import (
	"io"
)

// StreamHandler writes logs to somewhere
type StreamHandler interface {
	Write(p []byte) (n int, err error)
	Close() error
}

// StreamHandle writes logs to a specified io Writer, maybe stdout, stderr, etc...
type StreamHandle struct {
	w io.Writer
}

func NewStreamHandle(w io.Writer) (*StreamHandle, error) {
	h := new(StreamHandle)

	h.w = w

	return h, nil
}

func (h *StreamHandle) Write(b []byte) (n int, err error) {
	return h.w.Write(b)
}

func (h *StreamHandle) Close() error {
	return nil
}

// NullHandler does nothing, it discards anything.
type NullHandler struct {
}

func NewNullHandler() (*NullHandler, error) {
	return new(NullHandler), nil
}

func (h *NullHandler) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (h *NullHandler) Close() {

}
