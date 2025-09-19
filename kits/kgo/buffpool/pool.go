package buffpool

import (
	"bytes"
	"io"
	"sync"
)

var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func GetBytesBuffer() *bytes.Buffer {
	bf := bufferPool.Get().(*bytes.Buffer)
	bf.Reset()
	return bf
}

func PutBytesBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	bufferPool.Put(buf)
}

type closeReleaseBuff struct {
	*bytes.Buffer
}

func (c closeReleaseBuff) Close() error {
	PutBytesBuffer(c.Buffer)
	return nil
}

func ReleaseCloser(buffer *bytes.Buffer) io.ReadCloser {
	return closeReleaseBuff{Buffer: buffer}
}
