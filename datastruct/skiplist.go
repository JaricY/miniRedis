package datastruct

import "math/rand"

const (
	ZSKIPLIST_MAXLEVEL = 1
	ZSKIPLIST_P        = 0.25
)

type ZskipList struct {
	Header *ZskipListNode
	Tail   *ZskipListNode
	Length uint64
	Level  int32
}

type ZskipListNode struct {
	Ele      *SDS
	Score    float32
	Backward *ZskipListNode
	Level    []ZskipListLevel
}

type ZskipListLevel struct {
	Span    uint64 //跟下一个结点之间的跨度
	Forward *ZskipListNode
}

func ZslCreate() *ZskipList {
	header := ZslCreateNode(ZSKIPLIST_MAXLEVEL, 0, nil)
	for j := 0; j < ZSKIPLIST_MAXLEVEL; j++ {
		header.Level[j].Forward = nil
		header.Level[j].Span = 0
	}
	header.Backward = nil
	zsl := &ZskipList{
		Level:  1,
		Length: 0,
		Header: header,
		Tail:   nil,
	}

	return zsl
}

func ZslCreateNode(level int32, score float32, ele *SDS) *ZskipListNode {
	node := &ZskipListNode{
		Level: make([]ZskipListLevel, level),
		Score: score,
	}
	if ele != nil {
		node.Ele = ele
	}
	return node
}

func (zsl *ZskipList) ZslFree() {
	// TODO
}

func (zsl *ZskipList) ZslInsert(score float32, ele *SDS) *ZskipListNode {
	// insert流程：
	// 1.先确定在底层的插入位置(有序的，可用二分法)
	// 2.然后在上层建表中插入，根据一定的概率来决定插入到哪一层，并且更新对应层的前后指针
	// 3.要保证高度不超过logN(N是节点数)，对跳表进行调成

	update := make([]*ZskipListNode, ZSKIPLIST_MAXLEVEL)
	rank := make([]uint64, ZSKIPLIST_MAXLEVEL)

	x := zsl.Header
	for i := zsl.Level - 1; i >= 0; i-- {
		if i == zsl.Level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.Level[i].Forward != nil &&
			(x.Level[i].Forward.Score < score ||
				(x.Level[i].Forward.Score == score &&
					SDSCmp(*x.Level[i].Forward.Ele, *ele) < 0)) {
			rank[i] += x.Level[i].Span
			x = x.Level[i].Forward
		}

		update[i] = x
	}

	level := ZslRandomLevel() // 随机获取一层
	if level > zsl.Level {    //如果比当前的最高层都要高
		for i := zsl.Level; i < level; i++ {
			rank[i] = 0
			update[i] = zsl.Header
			update[i].Level[i].Span = zsl.Length
		}
		zsl.Level = level
	}
	x = ZslCreateNode(level, score, ele)
	for i := int32(0); i < level; i++ {
		x.Level[i].Forward = update[i].Level[i].Forward
		update[i].Level[i].Forward = x

		x.Level[i].Span = update[i].Level[i].Span - (rank[0] - rank[i])
		update[i].Level[i].Span = (rank[0] - rank[i]) + 1
	}

	//如果获取到的level不是最高的，则要对上层的所有前一个点的span加一
	for i := level; i < zsl.Level; i++ {
		update[i].Level[i].Span++
	}

	//更新新插入的节点在最底层的位置
	//判断是否是头结点
	if update[0] == zsl.Header {
		x.Backward = nil
	} else {
		x.Backward = update[0]
	}

	//判断是否是尾节点
	if x.Level[0].Forward != nil {
		x.Level[0].Forward.Backward = x
	} else {
		zsl.Tail = x
	}

	zsl.Length++
	return x
}

func ZslRandomLevel() int32 {
	level := 1
	for float64(rand.Int31()&0xFFFF) < (ZSKIPLIST_P*0xFFFF) && level < ZSKIPLIST_MAXLEVEL {
		level++
	}
	return int32(level)
}

func (zsl *ZskipList) ZslGetRank(score float32, ele *SDS) uint64 {
	x := zsl.Header
	var rank uint64
	for i := zsl.Level - 1; i >= 0; i-- {
		//从顶层向下查找元素
		for x.Level[i].Forward != nil &&
			(x.Level[i].Forward.Score < score ||
				(x.Level[i].Forward.Score == score &&
					SDSCmp(*x.Level[i].Forward.Ele, *ele) <= 0)) {
			rank += x.Level[i].Span
			x = x.Level[i].Forward
		}
		// 找到每层的最后一个小于当前结点的结点，并判断是否相同
		if x.Ele != nil && SDSCmp(*x.Ele, *ele) == 0 {
			return rank
		}
	}

	return 0
}

func (zsl *ZskipList) Print() {
	for i := zsl.Level - 1; i >= 0; i-- {
		x := zsl.Header

		for x.Level[i].Forward != nil {
			if x.Level[i].Forward.Ele.Buf == nil {
				continue
			}
			print(string(x.Ele.Buf))
			for j := uint64(0); j < x.Level[i].Forward.Level[i].Span; j++ {
				print("   -> ")
			}
			x = x.Level[i].Forward
		}
		println("")
	}
}
