package jnet

import "jamger/net/tcp"

func init() {

}

// ------------------------- outside -------------------------

func SetCallback(f func(id uint64, pack tcp.Pack)) {
	tcp.SetCallback(f)
}

func Run() {
	tcp.Run()
}
