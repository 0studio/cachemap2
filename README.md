# Demo
	m := NewUint64SafeCacheMap(time.Microsecond)
	now := time.Now()
	obj := NewCacheObject("hello", now, 1)
	m.Put(1, obj)
	cacheObj, ok := m.Get(1, now)
    fmt.Println(cacheObj.(string))
