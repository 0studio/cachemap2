package cachemap

import (
	"github.com/jonhoo/drwmutex"
	"math/rand"
	"sync"
)

type List []uint64

func (this List) Len() int {
	return len(this)
}

func (this List) Get(index int) uint64 {
	return this[index]
}

func (this List) Equal(index int, userId uint64) bool {
	if this[index] == userId {
		return true
	} else {
		return false
	}
}

func (this List) Less(index int, userId uint64) bool {
	if this[index] < userId {
		return true
	} else {
		return false
	}
}

type SortList struct {
	list  List
	mutex drwmutex.DRWMutex
}

func CreateSortList() SortList {
	return SortList{
		list:  make(List, 0),
		mutex: drwmutex.New(),
	}
}

func (this SortList) lock() {
	this.mutex.Lock()
}

func (this SortList) unlock() {
	this.mutex.Unlock()
}

func (this SortList) rLock() sync.Locker {
	return this.mutex.RLock()
}

func (this SortList) rUnlock(l sync.Locker) {
	l.Unlock()
}

func (this SortList) Get(index int) uint64 {
	l := this.rLock()
	defer this.rUnlock(l)
	return this.list.Get(index)
}

func (this SortList) Len() int {
	l := this.rLock()
	defer this.rUnlock(l)
	return this.list.Len()
}

func (this SortList) List() List {
	l := this.rLock()
	defer this.rUnlock(l)
	list := make(List, this.list.Len())
	copy(list, this.list)
	return list
}

func ListPush(sortList SortList, userId uint64) SortList {
	sortList.lock()
	defer sortList.unlock()
	length := sortList.list.Len()
	if length == 0 {
		sortList.list = []uint64{userId}
		return sortList
	}
	index, exist := listFind(sortList.list, userId)
	if exist {
		return sortList
	}
	newList := make(List, length+1)
	route := copy(newList[0:], sortList.list[:index])
	newList[route] = userId
	copy(newList[route+1:], sortList.list[index:])
	sortList.list = newList
	return sortList
}

func ListPop(sortList SortList, userId uint64) SortList {
	sortList.lock()
	defer sortList.unlock()
	length := sortList.list.Len()
	if length == 0 {
		return sortList
	} else if length == 1 {
		if sortList.list[0] == userId {
			sortList.list = []uint64{}
			return sortList
		} else {
			return sortList
		}
	}
	index, exist := listFind(sortList.list, userId)
	if !exist {
		return sortList
	}
	newList := make(List, length-1)
	roult := copy(newList[0:], sortList.list[:index])
	copy(newList[roult:], sortList.list[index+1:])
	sortList.list = newList
	return sortList
}

func ListPops(sortList SortList, idList []uint64) SortList {
	for _, id := range idList {
		sortList = ListPop(sortList, id)
	}
	return sortList
}

func RandCheckList(length, checkCnt int, isTotal bool) []int {
	var randList IntSlice
	randList = rand.Perm(length)
	if !isTotal && length > checkCnt {
		randList = randList[:checkCnt]
	}
	randList.Sort()
	return randList
}

func listFind(list List, userId uint64) (int, bool) {
	startPos := 0
	endPos := list.Len() - 1
	middlePos := (startPos + endPos) / 2
	for !list.Equal(middlePos, userId) && startPos < endPos {
		if list.Less(middlePos, userId) {
			startPos = middlePos + 1
		} else {
			endPos = middlePos - 1
		}
		middlePos = (startPos + endPos) / 2
	}
	var exist bool
	if list.Equal(middlePos, userId) {
		exist = true
	} else if list.Less(middlePos, userId) {
		middlePos++
	}
	return middlePos, exist
}
