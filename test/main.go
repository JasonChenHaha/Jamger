package main

import (
	"fmt"
	"jconfig"
	"jglobal"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

func main() {
	jconfig.Init()
	jglobal.Init()
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
	var a []int
	test(a...)
}

func test(b ...int) {
	fmt.Println(len(b))
}
