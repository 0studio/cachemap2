package cachemap

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCacheMap(t *testing.T) {
	m := make(Uint64CacheMap)
	now := time.Now()
	obj := NewCacheObject(100, now, 1)
	m.Put(1, obj)
	cacheObj, ok := m.Get(1, now)
	assert.True(t, ok)
	assert.True(t, 100 == cacheObj.(int))
	time.Sleep(1001 * time.Millisecond)
	cacheObj, ok = m.Get(1, time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get(2, time.Now())
	assert.False(t, ok)

}

func TestSafeCacheMap(t *testing.T) {
	m := NewUint64SafeCacheMap(time.Microsecond, callback, 4, 3)
	now := time.Now()
	obj := NewCacheObject(100, now, 1)
	m.Put(1, obj)
	assert.Equal(t, 1, m.Size())
	assert.Equal(t, 1, m.keyList.Len())
	cacheObj, ok := m.Get(1, now)
	assert.True(t, ok)
	assert.True(t, 100 == cacheObj.(int))
	assert.Equal(t, 1, m.Size())
	time.Sleep(1001 * time.Millisecond)
	assert.Equal(t, 0, m.Size()) // expires obj should be deleted
	assert.Equal(t, 0, m.keyList.Len())
	cacheObj, ok = m.Get(1, time.Now())
	assert.False(t, ok)

	cacheObj, ok = m.Get(2, time.Now())
	assert.False(t, ok)
}

func TestSafeMap(t *testing.T) {
	m := NewUint64SafeCacheMap(time.Millisecond*500, callback, 4, 3)
	now := time.Now()
	for i := 1; i < 1000; i++ {
		sec := i / 3
		obj := NewCacheObject(i, now, sec)
		m.Put(uint64(i), obj)
	}
	// time.Sleep(time.Second * 20)
	m.DropCallback()
}

func callback(object interface{}) {
	value := object.(int)
	fmt.Println("value:", value)
}
