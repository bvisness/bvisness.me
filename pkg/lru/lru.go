package lru

import (
	"hash/maphash"
	"sync"
	"time"
)

const bucketSize = 8

type LRU[V any] struct {
	bucketCount int
	list        []entry[V]
	buckets     []bucket[V]
	seed        maphash.Seed
}

type entry[V any] struct {
	set bool // if false, ignore this entry!

	key   string
	value V

	lastAccessed time.Time
}

type bucket[V any] struct {
	entries []entry[V] // subslice of the LRU's list
	m       sync.Mutex

	// stats!
	hits, misses      int
	stores, evictions int
}

func New[V any](size int) LRU[V] {
	if size%bucketSize != 0 {
		size += bucketSize - (size % bucketSize)
	}
	bucketCount := size / bucketSize

	list := make([]entry[V], size)
	buckets := make([]bucket[V], bucketCount)

	// initialize buckets
	for i := range buckets {
		bucketStart := bucketSize * i
		buckets[i].entries = list[bucketStart : bucketStart+bucketSize]
	}

	return LRU[V]{
		bucketCount: bucketCount,
		list:        list,
		buckets:     buckets,
		seed:        maphash.MakeSeed(),
	}
}

func (c *LRU[V]) bucketForKey(key string) *bucket[V] {
	var h maphash.Hash
	h.SetSeed(c.seed)
	h.WriteString(key)
	return &c.buckets[int(h.Sum64()%uint64(c.bucketCount))]
}

func (c *LRU[V]) Get(key string) (V, bool) {
	bucket := c.bucketForKey(key)
	bucket.m.Lock()
	defer bucket.m.Unlock()

	return c.bucketGet(bucket, key)
}

func (c *LRU[V]) bucketGet(bucket *bucket[V], key string) (V, bool) {
	for i := range bucket.entries {
		entry := &bucket.entries[i]
		if !entry.set {
			continue
		}
		if entry.key == key {
			bucket.hits++
			entry.lastAccessed = time.Now()
			return entry.value, true
		}
	}

	bucket.misses++
	var zero V
	return zero, false
}

func (c *LRU[V]) Store(key string, value V) {
	bucket := c.bucketForKey(key)
	bucket.m.Lock()
	defer bucket.m.Unlock()

	c.bucketStore(bucket, key, value)
}

func (c *LRU[V]) bucketStore(bucket *bucket[V], key string, value V) {
	// scan the bucket
	var firstEmpty, oldest *entry[V]
	for i := range bucket.entries {
		entry := &bucket.entries[i]
		if entry.set {
			if key == entry.key {
				// early-out: if this key is already stored, just update its entry
				bucket.stores++
				entry.value = value
				entry.lastAccessed = time.Now()
				return
			}
			if oldest == nil || entry.lastAccessed.Before(oldest.lastAccessed) {
				oldest = entry
			}
		} else if firstEmpty == nil {
			firstEmpty = entry
		}
	}

	// store in either the first empty slot or the oldest slot (overwriting it)
	storeAt := firstEmpty
	if storeAt == nil {
		bucket.evictions++
		storeAt = oldest
	}
	bucket.stores++
	*storeAt = entry[V]{
		set:          true,
		key:          key,
		value:        value,
		lastAccessed: time.Now(),
	}
}

func (c *LRU[V]) GetOrStore(key string, init func() (V, error)) (V, error) {
	bucket := c.bucketForKey(key)
	bucket.m.Lock()
	defer bucket.m.Unlock()

	if v, ok := c.bucketGet(bucket, key); ok {
		return v, nil
	} else {
		if v, err := init(); err == nil {
			c.bucketStore(bucket, key, v)
			return v, nil
		} else {
			var zero V
			return zero, err
		}
	}
}

type Stats struct {
	TotalHits, TotalMisses      int
	TotalStores, TotalEvictions int
}

func (c *LRU[V]) Stats() Stats {
	var stats Stats

	// lock all buckets
	for i := range c.buckets {
		bucket := &c.buckets[i]
		bucket.m.Lock()
		defer bucket.m.Unlock()

		stats.TotalHits += bucket.hits
		stats.TotalMisses += bucket.misses
		stats.TotalStores += bucket.stores
		stats.TotalEvictions += bucket.evictions
	}

	return stats
}
