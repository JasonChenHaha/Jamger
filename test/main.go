package main

import (
	"jlog"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

func main() {
	jlog.Info("<test start>")
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
	a := &ABC{}
	var e error
	e = a
	jlog.Debug(e.Error())
}

type ABC struct{}

func (a *ABC) Error() string {
	return "hello world"
}
