package sortedset

import "strconv"

type SortedSet struct {
	dict     map[string]*Element
	skiplist *skiplist
}

func Make() *SortedSet {
	return &SortedSet{
		dict:     make(map[string]*Element),
		skiplist: makeSkiplist(),
	}
}

// Add 返回值是说是否添加了一个新的值
func (sortedSet *SortedSet) Add(member string, score float64) bool {
	element, ok := sortedSet.dict[member]
	// 先更新dict，如果存在则直接更新，不存在则直接加入
	sortedSet.dict[member] = &Element{
		Member: member,
		Score:  score,
	}

	// 说明是旧值
	if ok {
		if score != element.Score {
			sortedSet.skiplist.remove(member, element.Score)
			sortedSet.skiplist.insert(member, score)
		}
		return false
	}
	sortedSet.skiplist.insert(member, score)
	return true
}

func (sortedSet *SortedSet) Len() int64 {
	return int64(len(sortedSet.dict))
}

func (sortedSet *SortedSet) Get(member string) (element *Element, ok bool) {
	element, ok = sortedSet.dict[member]
	if !ok {
		return nil, false
	}
	return element, true
}

func (sortedSet *SortedSet) Remove(member string) bool {
	v, ok := sortedSet.dict[member]
	if ok {
		sortedSet.skiplist.remove(member, v.Score)
		delete(sortedSet.dict, member)
		return true
	}
	return false
}

// GetRank 获取排名
func (sortedSet *SortedSet) GetRank(member string, desc bool) (rank int64) {
	element, ok := sortedSet.dict[member]
	if !ok {
		return -1
	}
	r := sortedSet.skiplist.getRank(member, element.Score)
	if desc {
		r = sortedSet.skiplist.length - r
	} else {
		r--
	}
	return r
}

func (sortedSet *SortedSet) ForEach(start int64, stop int64, desc bool, consumer func(element *Element) bool) {
	size := int64(sortedSet.Len())
	if start < 0 || start >= size {
		panic("illegal start " + strconv.FormatInt(start, 10))
	}
	if stop < start || stop > size {
		panic("illegal end " + strconv.FormatInt(stop, 10))
	}

	// 获取开始结点
	var node *node
	if desc {
		node = sortedSet.skiplist.tail
		if start > 0 {
			node = sortedSet.skiplist.getByRank(int64(size - start))
		}
	} else {
		node = sortedSet.skiplist.header.level[0].forward
		if start > 0 {
			node = sortedSet.skiplist.getByRank(int64(start + 1))
		}
	}

	sliceSize := int(stop - start)
	for i := 0; i < sliceSize; i++ {
		if !consumer(&node.Element) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
	}
}

// Range 返回排名在（start，stop）之间的元素，如果是正序则从0开始遍历
func (sortedSet *SortedSet) Range(start int64, stop int64, desc bool) []*Element {
	sliceSize := int(stop - start)
	slice := make([]*Element, sliceSize)
	i := 0
	sortedSet.ForEach(start, stop, desc, func(element *Element) bool {
		slice[i] = element
		i++
		return true
	})
	return slice
}

// Count 用于计算一个范围内的数量
func (sortedSet *SortedSet) Count(min *ScoreBorder, max *ScoreBorder) int64 {
	var i int64 = 0
	sortedSet.ForEach(0, sortedSet.Len(), false, func(element *Element) bool {
		gtMin := min.less(element.Score) // 判断左边界值小于当前值
		if !gtMin {
			// 说明当前值比最小值还小，继续遍历
			return true
		}
		ltMax := max.greater(element.Score) // 判断右边界值大于当前值
		if !ltMax {
			// 说明超过了最大值，停止遍历
			return false
		}
		// 说明当前值在最小值和最大值之间
		i++
		return true
	})
	return i
}

// ForEachByScore 用于遍历分数值在给定范围内的元素，并按照一定的偏移量和限制数目进行输出
func (sortedSet *SortedSet) ForEachByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool, consumer func(element *Element) bool) {
	// 找到开始遍历的节点
	var node *node
	if desc {
		node = sortedSet.skiplist.getLastInScoreRange(min, max)
	} else {
		node = sortedSet.skiplist.getFirstInScoreRange(min, max)
	}

	// 根据偏移量定位到指定位置开始遍历
	for node != nil && offset > 0 {
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		offset--
	}

	for i := 0; (i < int(limit) || limit < 0) && node != nil; i++ {
		if !consumer(&node.Element) { //访问每一个元素，并进行回调函数的处理
			break
		}
		if desc {
			node = node.backward //向前遍历
		} else {
			node = node.level[0].forward //向后遍历
		}
		if node == nil {
			break
		}
		gtMin := min.less(node.Element.Score)    //比较节点的分值是否大于最小值
		ltMax := max.greater(node.Element.Score) //比较节点的分值是否小于最大值
		if !gtMin || !ltMax {
			break //如果分值超出范围则退出循环
		}
	}
}

// RangeByScore returns members which score within the given border
// param limit: <0 means no limit
func (sortedSet *SortedSet) RangeByScore(min *ScoreBorder, max *ScoreBorder, offset int64, limit int64, desc bool) []*Element {
	if limit == 0 || offset < 0 {
		return make([]*Element, 0)
	}
	slice := make([]*Element, 0)
	sortedSet.ForEachByScore(min, max, offset, limit, desc, func(element *Element) bool {
		slice = append(slice, element)
		return true
	})
	return slice
}

func (sortedSet *SortedSet) RemoveByScore(min *ScoreBorder, max *ScoreBorder) int64 {
	removed := sortedSet.skiplist.RemoveRangeByScore(min, max, 0)
	for _, element := range removed {
		delete(sortedSet.dict, element.Member)
	}
	return int64(len(removed))
}

func (sortedSet *SortedSet) PopMin(count int) []*Element {
	first := sortedSet.skiplist.getFirstInScoreRange(negativeInfBorder, positiveInfBorder)
	if first == nil {
		return nil
	}
	border := &ScoreBorder{
		Value:   first.Score,
		Exclude: false,
	}
	removed := sortedSet.skiplist.RemoveRangeByScore(border, positiveInfBorder, count)
	for _, element := range removed {
		delete(sortedSet.dict, element.Member)
	}
	return removed
}

// RemoveByRank removes member ranking within [start, stop)
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) RemoveByRank(start int64, stop int64) int64 {
	removed := sortedSet.skiplist.RemoveRangeByRank(start+1, stop+1)
	for _, element := range removed {
		delete(sortedSet.dict, element.Member)
	}
	return int64(len(removed))
}
