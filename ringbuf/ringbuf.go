package ringbuf

import (
	"unsafe"
)

type ringBuf struct {
	keys   []string
	values []unsafe.Pointer
	qDel   queue
}

func (r *ringBuf) ClearExpired() {
	r.keys = r.keys[:0]
	r.values = r.values[:0]
	r.qDel.Clear()
}

func (r *ringBuf) Add(key string, data unsafe.Pointer) int32 {
	size := r.qDel.Pop()
	if size == r.qDel.NotFound() {
		r.values = append(r.values, data)
		r.keys = append(r.keys, key)
		return int32(len(r.values) - 1)
	}
	r.values[size] = data
	r.keys[size] = key
	return size
}

func (r *ringBuf) Delete(u int32) {
	r.qDel.Push(u)
}

func (r *ringBuf) Replace(u int32, data unsafe.Pointer) {
	r.values[u] = data
	return
}

func (r *ringBuf) CheckKey(idx int32, key string) bool {
	return r.keys[idx] == key
}

func (r *ringBuf) Get(u uint32) unsafe.Pointer {
	return r.values[u]
}

func newRingBuf() *ringBuf {
	return &ringBuf{}
}

type RingBuf struct {
	*ringBuf
}

func NewRingBuf() *RingBuf {
	return &RingBuf{
		ringBuf: newRingBuf(),
	}
}
