package cachemap

import (
	"time"
)

var UninitedTime time.Time

type CacheObject struct {
	expireTime time.Time
	obj        interface{}
}

func (cacheObj *CacheObject) Get() interface{} {
	return cacheObj.obj
}

func (cacheObj *CacheObject) GetObject(now time.Time) (obj interface{}, ok bool) {
	if cacheObj.expireTime == UninitedTime {
		ok = false
		return
	}
	if cacheObj.expireTime.After(now) {
		ok = true
		obj = cacheObj.obj
		return
	}
	return
}

func (cacheObj *CacheObject) UpdateObject(obj interface{}) bool {
	cacheObj.obj = obj
	return true
}

func (cacheObj *CacheObject) UpdateObjectWithExpireTime(obj interface{}, expireTime time.Time) bool {
	cacheObj.obj = obj
	cacheObj.expireTime = expireTime
	return true
}
func (cacheObj *CacheObject) UpdateObjectWithExpireTimeDur(obj interface{}, now time.Time, expireDur int) bool {
	cacheObj.obj = obj
	cacheObj.expireTime = now.Add(time.Duration(expireDur) * time.Second)
	return true
}

func NewCacheObject(obj interface{}, now time.Time, expireSeconds int) (cacheObj CacheObject) {
	cacheObj = CacheObject{
		expireTime: now.Add(time.Duration(expireSeconds) * time.Second),
		obj:        obj,
	}
	return
}

/***************************** callback runtime *****************************/
type Callback struct {
}
