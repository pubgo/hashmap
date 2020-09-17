package ringbuf

import (
	"math"
)

type queue struct {
	value []int32
}

func (q *queue) Clear() {
	q.value = q.value[:0]
}

// Len 长度判断
func (q *queue) Len() int {
	return len(q.value)
}

// Push ...
func (q *queue) Push(val int32) {
	q.value = append(q.value, val)
}

func (q *queue) NotFound() int32 {
	return math.MaxInt32
}

// Pop ...
func (q *queue) Pop() int32 {
	l := len(q.value)

	if l == 0 {
		return q.NotFound()
	}

	val := q.value[l-1]
	q.value = q.value[:l-1]
	return val
}
