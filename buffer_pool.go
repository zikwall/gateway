package gateway

import "sync"

type bufferPool struct {
	pool *sync.Pool
}

func newBufferPool(startSize int) *bufferPool {
	return &bufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &poolItem{
					data: make([]byte, 0, startSize),
				}
			},
		},
	}
}

// see https://staticcheck.io/docs/checks#SA6002
type poolItem struct {
	data []byte
}

func (b *bufferPool) Get() []byte {
	item, ok := b.pool.Get().(*poolItem)
	if ok {
		item.data = item.data[:0]
		return item.data
	}
	return nil
}

func (b *bufferPool) Put(buffer []byte) {
	b.pool.Put(&poolItem{
		data: buffer,
	})
}
