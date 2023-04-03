package main

import (
	"miniRedis/src"
)

func main() {
	sds := src.NewSDS("hello")
	println(sds.String())
}
