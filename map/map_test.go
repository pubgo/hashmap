package sync

import (
	"github.com/orcaman/concurrent-map"
	"github.com/pubgo/hashmap/internal"
	"strings"
	"sync"
	"testing"
	"time"
)

var ss = strings.Repeat("#", 1024)

func BenchmarkMap(b *testing.B) {
	dirty := make(map[string]string)
	var mu sync.Mutex
	for i := 0; i < b.N; i++ {
		mu.Lock()
		dirty[strings.Repeat("#", i%50)] = ss
		mu.Unlock()
	}
}

//func BenchmarkSyncMap(b *testing.B) {
//	var dirty sync.Map
//	go func() {
//		for {
//			for i := 0; i < b.N; i++ {
//				dirty.Load(strings.Repeat("#", i%50))
//			}
//			time.Sleep(time.Millisecond)
//		}
//	}()
//	for i := 0; i < b.N; i++ {
//		dirty.Store(strings.Repeat("#", i%50), ss)
//	}
//}

func BenchmarkSyncMap1(b *testing.B) {
	var dirty Map
	go func() {
		for {
			for i := 0; i < b.N; i++ {
				dirty.Load(strings.Repeat("#", i%50))
			}
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < b.N; i++ {
		dirty.Store(strings.Repeat("#", i%50), ss)
	}
}

func BenchmarkShardmap(b *testing.B) {
	dirty := cmap.New()
	go func() {
		for {
			for i := 0; i < b.N; i++ {
				dirty.Get(strings.Repeat("#", i%50))
			}
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < b.N; i++ {
		dirty.Set(strings.Repeat("#", i%50), ss)
	}
}

func BenchmarkShardmap1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		internal.MemHash([]byte(strings.Repeat("#", i%50)))
	}
}
