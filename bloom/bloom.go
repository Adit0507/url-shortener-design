package bloom

import (
	"fmt"
	"math"
	"sync"

	"github.com/cespare/xxhash/v2"
)

type BloomFilter struct {
	size      uint64
	hashCount uint
	bits      []bool
	mu        sync.RWMutex
}

func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	size := optimalSize(expectedItems, falsePositiveRate)
	hashCount := optimalHashCount(expectedItems, size)

	return &BloomFilter{
		size: uint64(size),
		hashCount: uint(hashCount),
		bits: make([]bool, size),
	}
}

// calculates optimal blom filter size
func optimalSize(n int, p float64) int {
	return int(-float64(n) * math.Log(p) / (math.Log(2) * math.Log(2)))
}

func optimalHashCount(n, m int) int {
	return int(math.Max(1, float64(m)/float64(n)*math.Log(2)))
}

func (bf *BloomFilter) Add(key string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := uint(0); i < bf.hashCount; i++ {
		hash := xxhash.Sum64([]byte(fmt.Sprintf("%s%d", key, i))) % bf.size
		bf.bits[hash] = true
	}
}

func (bf *BloomFilter) MightContain(key string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	for i := uint(0); i < bf.hashCount; i++ {
		hash := xxhash.Sum64([]byte(fmt.Sprintf("%s%d", key, i))) % bf.size
		if !bf.bits[hash] {
			return false
		}
	}

	return true
}
