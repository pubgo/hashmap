package hashmap

import (
	"bytes"
	"math/rand"
	"sync"
	"time"

	"github.com/pubgo/hashmap/internal"
	"github.com/pubgo/hashmap/ringbuf"
)

const defaultCap = 10
const factor = 6.5

type hashmap struct {
	cap       uint8
	slotsNum  uint32
	slotsNum1 uint32

	count  uint32
	count1 uint32

	rb    *ringbuf.RingBuf
	items [][]item
}

type item struct {
	hash  uint32
	index int32
}

func newHashmap() *hashmap {
	h := &hashmap{}
	h.cap = defaultCap
	h.slotsNum = 1<<h.cap - 1
	h.items = make([][]item, h.slotsNum+1)
	return h
}

func (h *hashmap) rehash(slot1 uint64) {
	if h.entities1 == nil {
		return
	}

	for h.entities1[slot1] != nil {
		ent := h.entities1[slot1]
		h.entities1[slot1] = ent.next
		h.count1--

		slot := internal.MemHash(ent.data[:ent.key]) & uint64(h.slotsNum)
		ent.next = h.entities[slot]
		h.entities[slot] = ent
		h.count++
	}
}

func (h *hashmap) rehash1() {
	if h.count1 > 0 {
		return
	}

	if h.entities1 != nil {
		h.entities1 = nil
	}

	if h.count > uint32(float64(h.slotsNum)*factor) {
		h.cap++
	} else if h.count < h.slotsNum && h.cap != defaultCap {
		h.cap--
	} else {
		return
	}

	h.slotsNum1 = h.slotsNum
	h.slotsNum = 1<<h.cap - 1

	h.entities1 = h.entities[:len(h.entities):len(h.entities)]
	h.entities = make([]*entity, h.slotsNum+1)

	h.count1 = h.count
	h.count = 0
}

func (h *hashmap) getSlots(key []byte) (uint64, uint64) {
	hk := internal.MemHash(key)
	return hk & uint64(h.slotsNum), hk & uint64(h.slotsNum1)
}

func (h *hashmap) get1(slot uint64, keyHash uint32, key string) (inx int32) {
	items := h.items[slot]
	for i := range items {
		if items[i].hash == keyHash && h.rb.CheckKey(items[i].index, key) {
			return items[i].index
		}
	}
	return
}

func (h *hashmap) get(key []byte) *item {
	sync.RWMutex{}
	var ent *entity
	slot, slot1 := h.getSlots(key)
	if h.entities1 != nil {
		ent = h.get1(h.entities1, slot1, key)
		// 迁移数据
	}

	if ent == nil {
		ent = h.get1(h.entities, slot, key)
	}
	return ent
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (h *hashmap) del1(entities []*entity, slot uint64, key []byte) *entity {
	var ent, pre *entity
	for ent = entities[slot]; ent != nil; ent = ent.next {
		if bytes.Equal(ent.data[:ent.key], key) {
			break
		}
		pre = ent
	}
	if ent == nil {
		return ent
	}

	if pre == nil {
		entities[slot] = ent.next
	} else {
		pre.next = ent.next
	}

	ent.next = h.delEntities
	h.delEntities = ent
	h.delNum++
	ent.data = ent.data[:0]
	h.rehash1()
	return ent
}

func (h *hashmap) del(key []byte) (ent *entity) {
	slot, slot1 := h.getSlots(key)
	if h.entities1 != nil {
		ent = h.del1(h.entities1, slot1, key)
		if ent != nil {
			h.count1--
		}
		h.rehash(slot1)
	}

	if ent == nil {
		ent = h.del1(h.entities, slot, key)
		if ent != nil {
			h.count--
		}
	}
	return
}

func (h *hashmap) lookup(slot []entryPtr, hash16 uint16, key []byte) (idx int, match bool) {
	idx = entryPtrIdx(slot, hash16)
	for idx < len(slot) {
		ptr := &slot[idx]
		if ptr.hash16 != hash16 {
			break
		}
		match = int(ptr.keyLen) == len(key) && seg.rb.EqualAt(key, ptr.offset+ENTRY_HDR_SIZE)
		if match {
			return
		}
		idx++
	}
	return
}

func (h *hashmap) set(key, val []byte) *entity {
	dl := len(key) + len(val)
	var dt = make([]byte, dl, dl)
	copy(dt[copy(dt, key):], val)

	var ent *entity
	slot, slot1 := h.getSlots(key)
	if h.entities1 != nil {
		ent = h.get1(h.entities1, slot1, key)
		h.rehash(slot1)
	}

	if ent == nil {
		ent = h.get1(h.entities, slot, key)
	}

	if ent == nil {
		if h.delEntities == nil {
			ent = &entity{key: uint8(len(key))}
		} else {
			ent = h.delEntities
			h.delEntities = h.delEntities.next
			h.delNum--
		}
		h.count++
	}

	ent.data = dt

	ent.next = h.entities[slot]
	h.entities[slot] = ent

	h.rehash1()
	return ent
}
