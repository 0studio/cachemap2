package cachemap

import (
	"github.com/jonhoo/drwmutex"
	"sync"
	"time"
)

const (
	CHECK_COUNT      = 16
	TOTAL_CLEAN_TIME = 3 // 凌晨3点
	CALLBACK_CHANNEL = 18
)

type CallbackFunc func(interface{})

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
type operateEnd struct {
	end chan struct{}
}

type Uint64SafeCacheMap struct {
	m                 Uint64CacheMap
	keyList           SortList
	cleanerTimer      chan bool
	callbackChan      chan interface{}
	callback          CallbackFunc
	autoCleanInterval time.Duration
	mutexList         []drwmutex.DRWMutex
}

// cleanTime -> total cleaning time(hour)
func NewUint64SafeCacheMap(autoCleanInterval time.Duration, callback CallbackFunc, funcCnt, cleanTime, mutexCnt int) (m *Uint64SafeCacheMap) {
	m = &Uint64SafeCacheMap{
		m:                 make(Uint64CacheMap),
		keyList:           CreateSortList(),
		cleanerTimer:      make(chan bool),
		callbackChan:      make(chan interface{}, CALLBACK_CHANNEL),
		callback:          callback,
		autoCleanInterval: autoCleanInterval,
	}
	for i := 0; i < mutexCnt; i++ {
		mutex := drwmutex.New()
		m.mutexList = append(m.mutexList, mutex)
	}
	for i := 0; i < funcCnt; i++ {
		go runCallback(m.callbackChan, m.callback)
	}
	go runTimer(m.cleanerTimer, m.autoCleanInterval, cleanTime)

	go func() {
		for {
			select {
			case total := <-m.cleanerTimer:
				m.process(total)
			}
		}
	}()
	return
}

func (this *Uint64SafeCacheMap) lock(key uint64) {
	index := key % uint64(len(this.mutexList))
	this.mutexList[index].Lock()
}

func (this *Uint64SafeCacheMap) unlock(key uint64) {
	index := key % uint64(len(this.mutexList))
	this.mutexList[index].Unlock()
}

func (this *Uint64SafeCacheMap) rLock(key uint64) sync.Locker {
	index := key % uint64(len(this.mutexList))
	return this.mutexList[index].RLock()
}

func (this *Uint64SafeCacheMap) rUnlock(l sync.Locker) {
	l.Unlock()
}

func (this *Uint64SafeCacheMap) Put(key uint64, obj CacheObject) {
	this.lock(key)
	defer this.unlock(key)
	this.m.Put(key, obj)
	this.keyList = ListPush(this.keyList, key)
}

func (this *Uint64SafeCacheMap) Get(key uint64, now time.Time) (obj interface{}, ok bool) {
	this.lock(key)
	defer this.unlock(key)
	this.checkKey(key, now)
	obj, ok = this.m.Get(key, now)
	return
}

func (this *Uint64SafeCacheMap) Delete(key uint64) bool {
	this.lock(key)
	defer this.unlock(key)
	ret, ok := this.m.Get(key, time.Now())
	if ok {
		this.m.Delete(key)
		this.keyList = ListPop(this.keyList, key)
		this.callbackChan <- ret
	}
	return true
}

func (this *Uint64SafeCacheMap) Size() int {
	return len(this.m)
}

func (this *Uint64SafeCacheMap) SaveAll() {
	for _, cacheObj := range this.m {
		obj := cacheObj.Get()
		saveFunc(this.callback, obj)
	}
}

func (this *Uint64SafeCacheMap) CheckKey(key uint64, now time.Time) {
	this.lock(key)
	defer this.unlock(key)
	this.checkKey(key, now)
}

func (this *Uint64SafeCacheMap) process(isTotal bool) {
	defer func() { recover() }()
	now := time.Now()
	randList := RandCheckList(this.keyList.Len(), CHECK_COUNT, isTotal)
	for _, index := range randList {
		ckId := this.keyList.Get(index)
		this.CheckKey(ckId, now)
	}
}

func (this *Uint64SafeCacheMap) checkKey(key uint64, now time.Time) {
	ret, ok := this.m.Expired(key, now)
	if ok {
		this.m.Delete(key)
		this.keyList = ListPop(this.keyList, key)
		this.callbackChan <- ret
	}
}

func runCallback(callbackChan <-chan interface{}, callbackFunc CallbackFunc) {
	for {
		select {
		case object := <-callbackChan:
			saveFunc(callbackFunc, object)
		}
	}
}

func runTimer(cleanerTimer chan<- bool, interval time.Duration, cleanTime int) {
	var isSweep, result bool = true, false
	for {
		time.Sleep(interval)
		if time.Now().Hour() > cleanTime {
			if !isSweep {
				isSweep, result = true, true
			} else {
				result = false
			}
		} else {
			isSweep, result = false, false
		}
		cleanerTimer <- result
	}
}

func saveFunc(fun func(interface{}), obj interface{}) {
	defer func() { recover() }()
	fun(obj)
}
