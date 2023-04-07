package datastruct

import (
	"fmt"
	"github.com/spaolacci/murmur3"
)

const (
	LONG_MAX                int64 = 2014
	DICT_HT_INIT_SIZE       int64 = 4
	DICT_FORCE_RESIZE_RATIO int64 = 1 //负载因子的阈值
	HASH_MAX_ZIPMAP_ENTRIES int16 = 512
)

type DictEntry struct {
	Key   *interface{}
	Value *interface{}
	Next  *DictEntry
}

type DictHt struct { // 哈希表

	Ht_table []*DictEntry // 哈希表数组
	Size     int64        // 哈希表大小
	SizeMask int64        // 哈希表大小掩码，总是等于 size -1
	Used     int64        // 已有结点数量

}

type Dict struct { //字典dict，哈希表实现，存储键值对
	DictType    DictType    //使用的字典类型
	Private     interface{} //私有数据,存储一些额外的信息
	Ht          [2]*DictHt  //哈希表
	Ht_Use      [2]int64    //已经被使用的哈希表节点数量
	RehashIdx   int         //正在进行rehash操作所对应的哈希表下标
	PauseRehash int16       //如果 > 0说明rehash被暂停了
	/*
		为什么要使用两个哈希表数组？
		哈希表是通过链地址法解决哈希冲突的，当链表过长的时候会导致查询的时间边长。
		1. 在进行扩容的时候是通过一次性创建一个更大的哈希表，然后将原哈希表中的所有键值对重新插入到新的哈表中
		2. 在这个过程中需要同时访问两个哈希表，原哈希表ht[0]和新哈希表ht[1]。
		3. 将原哈希表的内容全部复制到新哈希表之后，将ht[0]的空间释放并且将ht[1]设置为ht[0]。
		4. redis采用了读写所机制来保证rehash的线程安全性，即在rehash的过程中会获得写锁，
		   防止其他线程对哈希表进行修改，而在读取哈希表时，Redis 会获取读锁，允许多个线程同时读取哈希表。
		5. 在rehash的过程中，将新元素添加到ht[1]中，而不是添加到ht[0]中。
		   如果在复制的时候发现ht[1]中有该键值对，则直接删除Ht[0]中对应的键值对。
		6. 一般情况下是只使用ht[0]，ht[1]只在rehash中使用
	*/
}

type DictType interface { // 用于定义dict数据结构如何从操作
	HashFunction(key interface{}) int64     // 计算哈希值
	KeyDup(key interface{}) interface{}     // 复制key
	ValDup(key interface{}) interface{}     // 复制value
	KeyCompare(key1, key2 interface{}) bool // 比较key
	KeyDestructor(key interface{})          // 销毁key
	ValDestructor(key interface{})          // 销毁value
}

type MyDictType struct {
}

func (MyDictType) HashFunction(key interface{}) int64 {
	keyBytes := []byte(fmt.Sprint(key))
	h := murmur3.New64()
	h.Write(keyBytes)
	sum64 := h.Sum64()
	return int64(sum64)
}

func (MyDictType) KeyDup(key interface{}) interface{} {
	return nil
}

func (MyDictType) ValDup(key interface{}) interface{} {
	return nil
}
func (MyDictType) KeyCompare(key1, key2 interface{}) bool {
	key1Str := fmt.Sprint(key1)
	key2Str := fmt.Sprint(key2)
	if key1Str == key2Str {
		return true
	}
	return false
}
func (MyDictType) KeyDestructor(key interface{}) {

}
func (MyDictType) ValDestructor(key interface{}) {

}

func DictCreate(dictType DictType, privateData interface{}) *Dict {
	dict := &Dict{
		DictType: dictType,
		Private:  privateData,
	}
	dict.DictReset()
	return dict
}

func (d *Dict) DictReset() {
	d.Ht[0] = nil
	d.Ht[1] = nil
	d.RehashIdx = -1
	d.PauseRehash = 0
	d.Ht_Use[1] = 0
	d.Ht_Use[0] = 0
}

func (d *Dict) DictAdd(key, val interface{}) bool {

	var entry *DictEntry
	if d.dictIsRehashing() {
		// 继续rehash
		d.reHash()
	}
	d.dictExpandIfNeeded()

	ht := d.getHashTable() // 获取当前的hashTable

	hashKey := d.getHashKey(key) // 获取hashKey,是以值

	index := int64(ht.SizeMask) & hashKey // 获取到要插入的位置

	var prev *DictEntry = nil //从表头开始查找，判断是否存在相同的键

	curr := ht.Ht_table[index]

	for curr != nil {
		if *curr.Key == key { // 说明存在相同的键
			curr.Value = &val
			return false
		}
		prev = curr
		curr = curr.Next
	}

	// 走到这里说明没有相同的键
	entry = d.dictCreateEntry(&key, &val) //创建一个新的节点

	if prev == nil {
		ht.Ht_table[index] = entry
	} else {
		next := ht.Ht_table[index]
		entry.Next = next
		ht.Ht_table[index] = entry
	}

	ht.Used = ht.Used + 1

	d.Ht_Use[0]++

	// TODO 扩容判断

	return true
}

func (d *Dict) dictAddRaw(key interface{}, val interface{}, existing *DictEntry) *DictEntry {
	/**
	参数中的existing的主要原因是为了更灵活的控制
	当existing为空的时候，表示需要建立一个新的键值对插入字典中
	不为空则说明对一个已有的键值对进行更新操作
	*/

	return nil
}

func (d *Dict) dictIsRehashing() bool {
	return d.RehashIdx > -1
}

func (d *Dict) dictInitialized() bool {
	return d.DictType != nil
}

func (d *Dict) getHashTable() *DictHt {
	if d.dictIsRehashing() {
		return d.Ht[1]
	}
	return d.Ht[0]
}

func (d *Dict) getHashKey(key interface{}) int64 {
	return d.DictType.HashFunction(key)
}

func (d *Dict) dictCreateEntry(key, val *interface{}) *DictEntry {
	//将创建一个新的节点封装成函数
	entry := &DictEntry{
		Key:   key,
		Value: val,
		Next:  nil,
	}
	return entry
}

func (d *Dict) DictFetchValue(key interface{}) interface{} {
	entry := d.dictFind(&key)
	if entry != nil {
		return *entry.Value
	}
	return nil
}

func (d *Dict) dictFind(key *interface{}) *DictEntry {
	if d.dictIsRehashing() {
		d.reHash()
	}

	if d.dictGetSize() == 0 {
		return nil
	}

	for table := 0; table <= 1; table++ {
		ht := d.Ht[table] // 遍历两个hashtable
		if ht == nil {
			continue
		}
		hashKey := d.getHashKey(*key)
		index := ht.SizeMask & hashKey // 计算出索引下标
		entry := ht.Ht_table[index]
		for entry != nil {
			he_key := entry.Key
			if key == he_key || d.DictType.KeyCompare(*key, *he_key) { // key的地址相同或者key的值相同
				return entry
			}
			entry = entry.Next
		}
	}
	return nil

}

func (d *Dict) dictGetSize() int64 {
	return d.Ht_Use[0] + d.Ht_Use[1]
}

func (d *Dict) reHash() { //参数n表示每次处理多少个键值对
	n := HASH_MAX_ZIPMAP_ENTRIES
	if d.RehashIdx == -1 {
		d.RehashIdx = 0
	}

	ht0 := d.Ht[0]
	ht1 := d.Ht[1]

	for n > 0 && ht0.Used != 0 && int64(d.RehashIdx) < ht0.Size {
		dictEntry := ht0.Ht_table[d.RehashIdx]
		// 处理链表
		for dictEntry != nil {
			// 获取对应的下标
			index := d.getHashKey(*dictEntry.Key) & int64(ht1.SizeMask)
			// 获取下一个结点
			next := dictEntry.Next
			// 将当前的节点插入到链表的第一个
			dictEntry.Next = ht1.Ht_table[index]
			ht1.Ht_table[index] = dictEntry
			// 修改ht表
			ht0.Used--
			ht1.Used++

			dictEntry = next
		}
		//释放原表中的内容
		ht0.Ht_table[d.RehashIdx] = nil
		d.RehashIdx++

		// 如果所有键值对都映射完成，则rehash结束
		if int64(d.RehashIdx) == ht0.Size || ht0.Used == 0 {
			// 释放空间
			d.Ht[0] = ht1
			ht0.Ht_table = ht1.Ht_table
			d.Ht[1] = nil
			d.RehashIdx = -1
		}
		n--
	}
}
func (d *Dict) dictExpandIfNeeded() {
	// 判断是否需要扩容，如果需要则直接扩容
	// 如果当前正在库容，则不需要扩容
	if d.dictIsRehashing() {
		return
	}
	if d.Ht[0] == nil {
		// 初始化
		d.dictExpand(DICT_HT_INIT_SIZE)
	} else if d.Ht[0].Used >= d.Ht[0].Size {

		tmp := d.Ht[0].Size * DICT_FORCE_RESIZE_RATIO
		if d.Ht[0].Used >= tmp {
			//println("满足扩容条件，开始扩容")
			//println("扩容前use ->", d.Ht[0].Used)
			//println("扩容前size ->", d.Ht[0].Size)
			d.dictExpand(d.Ht[0].Size + 1)
			//println("扩容后use ->", d.Ht[0].Used)
			//println("扩容后size ->", d.Ht[0].Size)
		}
	}

}

func (d *Dict) dictExpand(size int64) {
	size = d.dictNextPower(size)
	ht := &DictHt{
		Ht_table: make([]*DictEntry, size),
		Size:     size,
		SizeMask: size - 1,
		Used:     0,
	}

	if d.Ht[0] == nil { // 说明是初始化扩容
		d.Ht[0] = ht
	} else if d.Ht[1] == nil || d.Ht[1].Ht_table == nil { // 如果正在进行rehash操作，则作为ht1.否则就是初始化库容
		d.Ht[1] = ht
		d.reHash()
	}

}

func (d *Dict) dictNextPower(size int64) int64 {
	if DICT_HT_INIT_SIZE == size {
		return DICT_HT_INIT_SIZE
	}
	if size > LONG_MAX {
		//如果到了边界
		return LONG_MAX + 1
	}

	i := DICT_HT_INIT_SIZE

	for i < size {
		i *= 2
	}

	return i
}

func (d *Dict) Print() {
	htTable := d.Ht[0].Ht_table
	for index := 0; index < len(htTable); index++ {
		entry := htTable[index]
		print(index, ":")
		if entry == nil {
			print("空")
		}
		for entry != nil {
			print("(", fmt.Sprint(*entry.Key), ",", fmt.Sprint(*entry.Value), ")->")
			entry = entry.Next
		}
		println("")
	}

	if d.Ht[1] != nil {
		println("ht1 start >>>>>>>>>>>>>>>>>>>>>>>>>")
		htTable := d.Ht[1].Ht_table
		for index := 0; index < len(htTable); index++ {
			entry := htTable[index]
			print(index, ":")
			if entry == nil {
				print("空")
			}
			for entry != nil {
				print("(", fmt.Sprint(*entry.Key), ",", fmt.Sprint(*entry.Value), ")->")
				entry = entry.Next
			}
			println("")
		}
		println("ht1 end >>>>>>>>>>>>>>>>>>>>>>>>>")
	}
}

func (d *Dict) DictReplace(key, val interface{}) {
	d.DictAdd(key, val)
}

func (d *Dict) DictRelease() {
	d.DictReset()
}

func (d *Dict) DictDelete(key interface{}) {
	for table := 0; table <= 1; table++ {
		ht := d.Ht[table] // 遍历两个hashtable
		if ht == nil {
			continue
		}
		hashKey := d.getHashKey(key)
		index := ht.SizeMask & hashKey // 计算出索引下标

		var pre *DictEntry
		pre = nil
		cur := ht.Ht_table[index]

		for cur != nil {
			he_key := cur.Key
			if &key == he_key || d.DictType.KeyCompare(key, *he_key) { // key的地址相同或者key的值相同
				if pre != nil {
					pre.Next = cur.Next
				} else {
					ht.Ht_table[index] = cur.Next
				}
				cur = nil
				ht.Used--
				return
			}
			pre = cur
			cur = cur.Next
		}
	}
}
