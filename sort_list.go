package cachemap

type SortList []uint64

func CreateSortList() SortList {
	return make(SortList, 0)
}

func (this SortList) Len() int {
	return len(this)
}

func (this SortList) Get(index int) uint64 {
	return this[index]
}

func (this SortList) Equal(index int, userId uint64) bool {
	if this[index] == userId {
		return true
	} else {
		return false
	}
}

func (this SortList) Less(index int, userId uint64) bool {
	if this[index] < userId {
		return true
	} else {
		return false
	}
}

func ListPush(list SortList, userId uint64) SortList {
	length := len(list)
	if length == 0 {
		return []uint64{userId}
	}
	index, exist := listFind(list, userId)
	if exist {
		return list
	}
	newList := make(SortList, length+1)
	route := copy(newList[0:], list[:index])
	newList[route] = userId
	copy(newList[route+1:], list[index:])
	return newList
}

func ListPop(list SortList, userId uint64) SortList {
	length := len(list)
	if length == 0 {
		return []uint64{}
	} else if length == 1 {
		if list[0] == userId {
			return []uint64{}
		} else {
			return list
		}
	}
	index, exist := listFind(list, userId)
	if !exist {
		return list
	}
	newList := make(SortList, length-1)
	roult := copy(newList[0:], list[:index])
	copy(newList[roult:], list[index+1:])
	return newList
}

func ListPops(list SortList, idList []uint64) SortList {
	for _, id := range idList {
		list = ListPop(list, id)
	}
	return list
}

func listFind(list SortList, userId uint64) (int, bool) {
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
