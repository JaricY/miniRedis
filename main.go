package main

import (
	"fmt"
	"miniRedis/src"
	"strconv"
)

func main() {
	dictType := &src.MyDictType{}
	dict := src.DictCreate(dictType, nil)
	for i := 1; i <= 33; i++ {
		add := dict.DictAdd("key"+strconv.Itoa(i), "val"+strconv.Itoa(i))
		if add {
			//println("第"+strconv.Itoa(i), "次添加结果")
			//dict.Print()
			//println("------------")
		}
		//time.Sleep(500 * time.Millisecond)
	}
	value := dict.DictFetchValue("key1")
	println(fmt.Sprint("before...", value))
	dict.DictReplace("key1", "newval1")
	value = dict.DictFetchValue("key1")
	println(fmt.Sprint("after...", value))
	dict.DictDelete("key1")
	value = dict.DictFetchValue("key1")
	println(fmt.Sprint("delete...", value))

}
