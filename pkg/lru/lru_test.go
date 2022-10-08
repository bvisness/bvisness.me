package lru

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {
	t.Run("sanity", func(t *testing.T) {
		size := 10
		c := New[int](size)

		realSize := len(c.list)
		t.Log("real size:", realSize)
		assert.GreaterOrEqual(t, realSize, size)
		assert.Equal(t, 0, realSize%bucketSize)

		_, ok := c.Get("foo")
		assert.False(t, ok)
		assert.Equal(t, 1, c.Stats().TotalMisses)

		c.Store("foo", 42)
		{
			v, ok := c.Get("foo")
			assert.True(t, ok)
			assert.Equal(t, 42, v)
			assert.Equal(t, 1, c.Stats().TotalHits)
			assert.Equal(t, 1, c.Stats().TotalMisses)
		}

		c.Store("foo", 42)
		{
			v, ok := c.Get("foo")
			assert.True(t, ok)
			assert.Equal(t, 42, v)
			assert.Equal(t, 2, c.Stats().TotalHits)
			assert.Equal(t, 1, c.Stats().TotalMisses)
		}
	})
	t.Run("single-thread evictions", func(t *testing.T) {
		size := bucketSize * 4 // 4 buckets
		c := New[int](size)

		// insert enough things to definitely evict entries
		numInserts := size * 100
		for i := 0; i < numInserts; i++ {
			key := fmt.Sprintf("%d", i)
			in := rand.Int()

			_, ok := c.Get(key) // deliberately miss once
			assert.False(t, ok)

			c.Store(key, in) // then store the new thing

			out, ok := c.Get(key) // then get it again (should hit)
			assert.True(t, ok)
			assert.Equal(t, in, out)
		}

		stats := c.Stats()
		assert.Equal(t, numInserts, stats.TotalMisses)
		assert.Equal(t, numInserts, stats.TotalStores)
		assert.Equal(t, numInserts, stats.TotalHits)
		assert.Equal(t, numInserts-size, stats.TotalEvictions)
	})
	t.Run("multi-threaded evictions", func(t *testing.T) {
		size := bucketSize * 4 // 4 buckets
		c := New[int](size)

		// insert enough things to definitely evict entries
		insertsPerWorker := size * 100
		numWorkers := 4
		numInserts := numWorkers * insertsPerWorker

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerId int) {
				for i := 0; i < insertsPerWorker; i++ {
					key := fmt.Sprintf("%d-%d", workerId, i)
					in := rand.Int()

					_, ok := c.Get(key) // deliberately miss once
					assert.False(t, ok)

					c.Store(key, in) // then store the new thing

					out, ok := c.Get(key) // then get it again (should hit)
					if ok {
						// It might have been evicted in between the store and get.
						assert.Equal(t, in, out, "got unexpected value from stored entry")
					}
				}
				wg.Done()
			}(i)
		}

		wg.Wait()

		stats := c.Stats()
		// stores and evictions are predictable...
		assert.Equal(t, numInserts, stats.TotalStores)
		assert.Equal(t, numInserts-size, stats.TotalEvictions)
		// but hits and misses are not. See if they're close.
		assertWithinPercentage(t, stats.TotalHits, numInserts, 0.1)
		assertWithinPercentage(t, stats.TotalMisses, numInserts, 0.1)
		t.Logf("%+v", stats)
	})
}

func assertWithinPercentage(t *testing.T, actual, expected int, pct float32) bool {
	t.Helper()

	allowedSlop := int(float32(expected) * pct)
	actualSlop := int(math.Abs(float64(actual - expected)))
	return assert.Less(t, actualSlop, allowedSlop, "%v was not within %v%% of %v (off by %v, must be off by less than %v)", actual, pct*100, expected, actualSlop, allowedSlop)
}
