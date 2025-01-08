package main

import (
	"fmt"
)

const (
	gHeadSize = 2
	gCmdSize  = 2
)

type A struct {
	a int
}

func (a *A) Add() {
	a.a++
}

func (a *A) Show() {
	fmt.Print(a.a)
}

type kkk interface {
	Add()
	Show()
}

func main() {
	// jconfig.Init()
	// jglobal.Init()
	// jlog.Init("")
	// testTcp()
	// testKcp()
	// testWeb()
	// testHttp()
	a := 0
	fmt.Println(GetDigitCnt(a))
}

func GetDigitCnt(n int) int {
	c := 0
	for n != 0 {
		c++
		n /= 10
	}
	return c
}
