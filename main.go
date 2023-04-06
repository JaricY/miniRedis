package main

import (
	"miniRedis/datastruct"
	"strconv"
)

func main() {
	zskipList := datastruct.ZslCreate()
	j := 1
	var i float32 = 0.1
	for ; i <= 0.9; i += 0.1 {
		sds := datastruct.NewSDS("key" + strconv.Itoa(j))
		zskipList.ZslInsert(i, sds)
		rank := zskipList.ZslGetRank(i, sds)
		println("rank", j, " ->", rank)
		j++
	}
	zskipList.Print()

}
