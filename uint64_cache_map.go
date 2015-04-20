package cachemap

import (
	"math/rand"
	"time"
)

const (
	CHECK_COUNT      = 16
	TOTAL_CLEAN_TIME = 3 // 凌晨3点
	CALLBACK_CHANNEL = 18
)

////////////////////////////////////////////////////////////////////////////////
type Uint64CacheMap map[uint64]CacheObject

func (m Uint64CacheMap) Put(key uint64, obj CacheObject) {
	m[key] = obj
}

func (m Uint64CacheMap) Get(key uint64, now time.Time) (obj interface{}, ok bool) {
	var cacheObj CacheObject
	cacheObj, ok = m[key]
	if !ok {
		return
	}
	obj, ok = cacheObj.GetObject(now)
	if !ok { // expired
		m.Delete(key)
	}
	return
}

func (m Uint64CacheMap) Expired(key uint64, now time.Time) (interface{}, bool) {
	cacheObj, ok := m[key]
	if !ok {
		return nil, false
	}
	_, ok = cacheObj.GetObject(now)
	if !ok {
		return cacheObj.Get(), true
	}
	return nil, false
}

func (m Uint64CacheMap) Delete(key uint64) bool {
	delete(m, key)
	return true
}

type uint64CacheObjectWapper struct {
	obj CacheObject
	key uint64
}
type resultWapper struct {
	obj interface{}
	ok  bool
}
type resultGetter struct {
	key    uint64
	now    time.Time
	result chan resultWapper
}
type sizeGetter struct {
	size chan int
}
type Uint64SafeCacheMap struct {
	m                 Uint64CacheMap
	keyList           SortList
	setChan           chan uint64CacheObjectWapper
	getChan           chan resultGetter
	delChan           chan uint64
	cleanerTimer      chan bool
	sizeChan          chan sizeGetter
	callbackChan      chan interface{}
	autoCleanInterval time.Duration
}

func NewUint64SafeCacheMap(autoCleanInterval time.Duration, callback func(interface{})) (m *Uint64SafeCacheMap) {
	m = &Uint64SafeCacheMap{
		m:                 make(Uint64CacheMap),
		keyList:           CreateSortList(),
		setChan:           make(chan uint64CacheObjectWapper),
		getChan:           make(chan resultGetter),
		delChan:           make(chan uint64),
		cleanerTimer:      make(chan bool),
		sizeChan:          make(chan sizeGetter),
		callbackChan:      make(chan interface{}, CALLBACK_CHANNEL),
		autoCleanInterval: autoCleanInterval,
	}
	go func() {
		for {
			m.process()
		}
	}()
	go runCallback(m.callbackChan, callback)
	go runTimer(m.cleanerTimer, m.autoCleanInterval)
	return
}
func (safeMap *Uint64SafeCacheMap) Put(key uint64, obj CacheObject) {
	safeMap.setChan <- uint64CacheObjectWapper{key: key, obj: obj}
}
func (safeMap *Uint64SafeCacheMap) Get(key uint64, now time.Time) (obj interface{}, ok bool) {
	getter := resultGetter{key: key, now: now, result: make(chan resultWapper)}
	safeMap.getChan <- getter
	result := <-getter.result
	obj = result.obj
	ok = result.ok
	close(getter.result)
	return
}

func (safeMap *Uint64SafeCacheMap) Size() (size int) {
	sizeGetter := sizeGetter{size: make(chan int)}
	safeMap.sizeChan <- sizeGetter
	size = <-sizeGetter.size
	close(sizeGetter.size)
	return
}

// 这个不会调用销毁回调函数
func (safeMap *Uint64SafeCacheMap) Delete(key uint64) bool {
	safeMap.delChan <- key
	return true
}
func (safeMap *Uint64SafeCacheMap) GetDirtyKeys() (keys []uint64) {
	keys = make([]uint64, 0, len(safeMap.m))
	for key, _ := range safeMap.m {
		keys = append(keys, key)
	}
	return
}

func (m *Uint64SafeCacheMap) process() {
	defer recover()
	select {
	case setter := <-m.setChan:
		m.m.Put(setter.key, setter.obj)
		m.keyList = ListPush(m.keyList, setter.key)
	case getter := <-m.getChan:
		m.checkKey(getter.key, getter.now)
		ret, ok := m.m.Get(getter.key, getter.now)
		getter.result <- resultWapper{obj: ret, ok: ok}
	case delId := <-m.delChan:
		m.m.Delete(delId)
		m.keyList = ListPop(m.keyList, delId)
	case total := <-m.cleanerTimer:
		now := time.Now()
		randList := randCheckList(m.keyList.Len(), total)
		for _, index := range randList {
			ckId := m.keyList.Get(index)
			m.checkKey(ckId, now)
		}
	case sizeGetter := <-m.sizeChan:
		sizeGetter.size <- len(m.m)
	}
}

func (m *Uint64SafeCacheMap) checkKey(key uint64, now time.Time) {
	ret, ok := m.m.Expired(key, now)
	if ok {
		m.m.Delete(key)
		m.keyList = ListPop(m.keyList, key)
		m.callbackChan <- ret
	}
}

func runCallback(callbackChan <-chan interface{}, callbackFunc func(interface{})) {
	for {
		select {
		case object := <-callbackChan:
			func() {
				defer recover()
				callbackFunc(object)
			}()
		}
	}
}

func runTimer(cleanerTimer chan<- bool, interval time.Duration) {
	var isSweep bool = true
	for {
		time.Sleep(interval)
		hour := time.Now().Hour()
		if hour < TOTAL_CLEAN_TIME {
			isSweep = false
			cleanerTimer <- false
		} else {
			if !isSweep {
				isSweep = true
				cleanerTimer <- true
			} else {
				cleanerTimer <- false
			}
		}
	}
}

func randCheckList(length int, isTotal bool) []int {
	var randList IntSlice
	randList = rand.Perm(length)
	if !isTotal && length > CHECK_COUNT {
		randList = randList[:CHECK_COUNT]
	}
	randList.Sort()
	return randList
}
