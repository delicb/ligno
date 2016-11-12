package ligno

import (
	"sync/atomic"
	"unsafe"
)

// replaceableHandler wraps another handler that may be swapped out
// dynamically at runtime in a thread-safe fashion.
// idea for replaceable handler is from log15 project.
type replaceableHandler struct {
	handler unsafe.Pointer
}

// Handler returns current handler.
func (h *replaceableHandler) Handler() Handler {
	return (*(*Handler)(atomic.LoadPointer(&h.handler)))
}

// Handle is implementation of Handler interface. It only passes record
// to underlying handler if it is set.
func (h *replaceableHandler) Handle(r Record) error {
	handler := h.Handler()
	if handler != nil {
		return handler.Handle(r)
	}
	return nil
}

// Close is implementation of HandlerCloser interface.
// If underlying handler implements HandlerCloser interface, its Close
// method will be called.
func (h *replaceableHandler) Close() {
	if closableHandler, ok := h.Handler().(HandlerCloser); ok {
		closableHandler.Close()
	}
}

// Replace sets provided handler as new underlying handler.
func (h *replaceableHandler) Replace(newHandler Handler) {
	atomic.StorePointer(&h.handler, unsafe.Pointer(&newHandler))
}
