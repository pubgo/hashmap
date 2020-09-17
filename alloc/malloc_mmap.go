package alloc

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const chunksPerAlloc = 1024
const chunkSize = 16

var (
	freeChunks     []*[chunkSize]byte
	freeChunksLock sync.Mutex
)

func getChunk() []byte {
	freeChunksLock.Lock()
	if len(freeChunks) == 0 {
		// Allocate offheap memory, so GOGC won't take into account cache size.
		// This should reduce free memory waste.
		data, err := syscall.Mmap(-1, 0, chunkSize*chunksPerAlloc, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			panic(fmt.Errorf("cannot allocate %d bytes via mmap: %s", chunkSize*chunksPerAlloc, err))
		}
		for len(data) > 0 {
			p := (*[chunkSize]byte)(unsafe.Pointer(&data[0]))
			freeChunks = append(freeChunks, p)
			data = data[chunkSize:]
		}
	}
	n := len(freeChunks) - 1
	p := freeChunks[n]
	freeChunks[n] = nil
	freeChunks = freeChunks[:n]
	freeChunksLock.Unlock()
	return p[:]
}

func putChunk(chunk []byte) {
	if chunk == nil {
		return
	}

	chunk = chunk[:chunkSize]
	p := (*[chunkSize]byte)(unsafe.Pointer(&chunk[0]))

	freeChunksLock.Lock()
	freeChunks = append(freeChunks, p)
	freeChunksLock.Unlock()
}

type ShmSpan struct {
	origin []byte
	name   string

	data   uintptr
	offset int
	size   int
}

func NewShmSpan(name string, data []byte) *ShmSpan {
	return &ShmSpan{
		name:   name,
		origin: data,
		data:   uintptr(unsafe.Pointer(&data[0])),
		size:   len(data),
	}
}

func (s *ShmSpan) Alloc(size int) (uintptr, error) {
	if s.offset+size > s.size {
		return 0, errNotEnough
	}

	ptr := s.data + uintptr(s.offset)
	s.offset += size
	return ptr, nil
}

func (s *ShmSpan) Data() uintptr {
	return s.data
}

func (s *ShmSpan) Origin() []byte {
	return s.origin
}

var (
	errNotEnough = errors.New("span capacity is not enough")
)

func checkConsistency(path string, size int) error {
	if info, err := os.Stat(path); err == nil {
		if info.Size() != int64(size) {
			return errors.New(fmt.Sprintf("mmap target path %s exists and its size %d mismatch %d", path, info.Size(), size))
		}
	}
	return nil
}

func Alloc(name string, size int) (*ShmSpan, error) {
	path := name

	os.MkdirAll(filepath.Dir(path), 0755)

	// check consistency
	if err := checkConsistency(path, size); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	if err := f.Truncate(int64(size)); err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {
		return nil, err
	}

	// lock mmap data to avoid I/O page fault
	err = syscall.Mlock(data)
	if err != nil {
		log.Printf("failed to mlock memory from mmap, please check the RLIMIT_MEMLOCK:%s\n", err)
	}

	return NewShmSpan(name, data), nil
}

func Free(span *ShmSpan) error {
	Clear(span.name)
	return syscall.Munmap(span.origin)
}

func Clear(name string) error {
	return os.Remove(name)
}
