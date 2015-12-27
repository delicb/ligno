package ligno

import (
	"bytes"
	"sync"
)

// byteBufferPool is pool of byte buffers, used to avoid allocation for each
// formatting of log message in formatter.
type byteBufferPool struct {
	sync.Pool
}

// newByteBufferPool creates new instance of byteBufferPool initialized
// to creates new buffer instances or initial size 128 bytes.
func newByteBufferPool() *byteBufferPool {
	return &byteBufferPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 128))
			},
		},
	}
}

// Get returns fresh buffer from pool.
func (bp *byteBufferPool) Get() *bytes.Buffer {
	buff := bp.Pool.Get().(*bytes.Buffer)
	buff.Reset()
	return buff
}

// Put returns buffer to pool, for later use.
func (bp *byteBufferPool) Put(buff *bytes.Buffer) {
	bp.Pool.Put(buff)
}

// buffPool is single instance of buffer pool.
var buffPool = newByteBufferPool()
