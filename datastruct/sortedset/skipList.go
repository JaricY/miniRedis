package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	maxLevel = 16
)

// Element 保存了元素的内容和分值
type Element struct {
	Member string
	Score  float64
}

// Level 表示层级 ，相当于是zskiplistLevel结构体
type Level struct {
	// 前驱指针
	forward *node
	// 与前一个点的跨度
	span int64
}

// 表示一个结点，相当于是zskiplistNode
type node struct {
	// 元素值和分数
	Element
	// 后驱指针
	backward *node
	level    []*Level // level[0] 是最底层
}

// 跳表结构
type skiplist struct {
	header *node
	tail   *node
	// 具有的元素个数
	length int64
	// 最高层级
	level int16
}

func makeNode(level int16, score float64, member string) *node {
	n := &node{
		Element: Element{
			Score:  score,
			Member: member,
		},
		level: make([]*Level, level),
	}
	for i := range n.level {
		n.level[i] = new(Level)
	}
	return n
}

func makeSkiplist() *skiplist {
	return &skiplist{
		level:  1,
		header: makeNode(maxLevel, 0, ""),
	}
}

func randomLevel() int16 {
	total := uint64(1)<<uint64(maxLevel) - 1
	k := rand.Uint64() % total
	return maxLevel - int16(bits.Len64(k+1)) + 1
}

// 跟源码是一样的
func (skiplist *skiplist) insert(member string, score float64) *node {
	update := make([]*node, maxLevel) // 用于存储需要更新的所有节点
	rank := make([]int64, maxLevel)   // 用于存储新的排名

	// 查找插入的位置
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		if i == skiplist.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1] // 存储在这一层级的排名rank
		}
		if node.level[i] != nil {
			// 遍历所有层级，存储最后一个小于需要插入值的结点
			for node.level[i].forward != nil &&
				(node.level[i].forward.Score < score ||
					(node.level[i].forward.Score == score && node.level[i].forward.Member < member)) { // same score, different key
				rank[i] += node.level[i].span
				node = node.level[i].forward
			}
		}
		update[i] = node
	}

	level := randomLevel()
	// 修改skiplist的最高层级情况
	if level > skiplist.level {
		for i := skiplist.level; i < level; i++ {
			rank[i] = 0
			update[i] = skiplist.header
			update[i].level[i].span = skiplist.length
		}
		skiplist.level = level
	}

	// 创建node插入到skiplist中
	node = makeNode(level, score, member)
	for i := int16(0); i < level; i++ {
		node.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = node

		// 通过rank更新span
		node.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	// 将上层的span的值都加一
	for i := level; i < skiplist.level; i++ {
		update[i].level[i].span++
	}

	// 设置前后结点
	if update[0] == skiplist.header {
		node.backward = nil
	} else {
		node.backward = update[0]
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node
	} else {
		skiplist.tail = node
	}
	skiplist.length++
	return node
}

func (skiplist *skiplist) removeNode(node *node, update []*node) {

	// 更新每层的节点
	for i := int16(0); i < skiplist.level; i++ {
		if update[i].level[i].forward == node {
			update[i].level[i].span += node.level[i].span - 1
			update[i].level[i].forward = node.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}
	// 更新删除后结点的前后结点
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		skiplist.tail = node.backward
	}

	// 更新skiplist的最高层级
	for skiplist.level > 1 && skiplist.header.level[skiplist.level-1].forward == nil {
		skiplist.level--
	}
	skiplist.length--
}

// 返回值是是否移除成功
func (skiplist *skiplist) remove(member string, score float64) bool {
	/*
	 * 和插入类似，查找到每层需要修改的点，也就是比score小的点
	 */
	update := make([]*node, maxLevel)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.Score < score ||
				(node.level[i].forward.Score == score &&
					node.level[i].forward.Member < member)) {
			node = node.level[i].forward
		}
		update[i] = node
	}

	node = node.level[0].forward
	// 判断这个点是否真的是需要删除的
	if node != nil && score == node.Score && node.Member == member {
		skiplist.removeNode(node, update)
		return true
	}
	return false
}

func (skiplist *skiplist) getRank(member string, score float64) int64 {
	// 相当于是遍历最底层的内容，但是如果使用Rank可以降低遍历的节点数量
	var rank int64 = 0
	x := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			(x.level[i].forward.Score < score ||
				(x.level[i].forward.Score == score &&
					x.level[i].forward.Member <= member)) {
			rank += x.level[i].span
			x = x.level[i].forward
		}

		/* x有可能是头结点，所以要确保不是null */
		if x.Member == member {
			return rank
		}
	}
	return 0
}

func (skiplist *skiplist) getByRank(rank int64) *node {
	var i int64 = 0
	n := skiplist.header
	// 扫描每层
	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && (i+n.level[level].span) <= rank {
			i += n.level[level].span
			n = n.level[level].forward
		}
		if i == rank {
			return n
		}
	}
	return nil
}

// hasInRange 判断是否在这个范围内
func (skiplist *skiplist) hasInRange(min *ScoreBorder, max *ScoreBorder) bool {
	// min & max = empty
	if min.Value > max.Value || (min.Value == max.Value && (min.Exclude || max.Exclude)) {
		return false
	}
	// min > tail
	n := skiplist.tail
	if n == nil || !min.less(n.Score) {
		return false
	}
	// max < head
	n = skiplist.header.level[0].forward
	if n == nil || !max.greater(n.Score) {
		return false
	}
	return true
}

func (skiplist *skiplist) getFirstInScoreRange(min *ScoreBorder, max *ScoreBorder) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	// 从顶层开始查找，找到第一个小于max大于min的值
	for level := skiplist.level - 1; level >= 0; level-- {
		// 遍历每层
		for n.level[level].forward != nil && !min.less(n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	/* 这是一个范围查询，所以下一个点不能是null. */
	n = n.level[0].forward
	if !max.greater(n.Score) {
		return nil
	}
	return n
}

func (skiplist *skiplist) getLastInScoreRange(min *ScoreBorder, max *ScoreBorder) *node {
	if !skiplist.hasInRange(min, max) {
		return nil
	}
	n := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for n.level[level].forward != nil && max.greater(n.level[level].forward.Score) {
			n = n.level[level].forward
		}
	}
	if !min.less(n.Score) {
		return nil
	}
	return n
}

// RemoveRangeByScore 根据排名移除limit个在（min，max）的元素
func (skiplist *skiplist) RemoveRangeByScore(min *ScoreBorder, max *ScoreBorder, limit int) (removed []*Element) {
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)
	// find backward nodes (of target range) or last node of each level
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil {
			if min.less(node.level[i].forward.Score) { // already in range
				break
			}
			node = node.level[i].forward
		}
		update[i] = node
	}

	// 取第一个在范围内的node
	node = node.level[0].forward

	//移除结点
	for node != nil {
		if !max.greater(node.Score) { // already out of range
			break
		}
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		skiplist.removeNode(node, update)
		if limit > 0 && len(removed) == limit {
			break
		}
		node = next
	}
	return removed
}

func (skiplist *skiplist) RemoveRangeByRank(start int64, stop int64) (removed []*Element) {
	var i int64 = 0 // rank of iterator
	update := make([]*node, maxLevel)
	removed = make([]*Element, 0)

	// scan from top level
	node := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for node.level[level].forward != nil && (i+node.level[level].span) < start {
			i += node.level[level].span
			node = node.level[level].forward
		}
		update[level] = node
	}

	i++
	node = node.level[0].forward // first node in range

	// remove nodes in range
	for node != nil && i < stop {
		next := node.level[0].forward
		removedElement := node.Element
		removed = append(removed, &removedElement)
		skiplist.removeNode(node, update)
		node = next
		i++
	}
	return removed
}
