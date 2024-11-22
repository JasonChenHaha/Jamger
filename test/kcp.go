package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jdebug "jamger/debug"
	jglobal "jamger/global"
	jlog "jamger/log"
	jkcp "jamger/net/kcp"
	"strings"
	"time"

	"github.com/xtaci/kcp-go"
)

type Kcp struct {
	con *kcp.UDPSession
}

func testKcp() *Kcp {
	jlog.Info("<test kcp>")
	kc := &Kcp{}
	addr := strings.Split(jconfig.GetString("kcp.addr"), ":")
	con, _ := kcp.DialWithOptions("127.0.0.1:"+addr[1], nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	jlog.Info("connect to server ", addr)
	kc.con = con

	go kc.heartbeat()

	kc.send(jglobal.CMD_PING, []byte{})
	kc.recv()

	return kc
}

func (kc *Kcp) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		kc.send(jglobal.CMD_HEARTBEAT, []byte{})
	}
}

func (kc *Kcp) close() {
	kc.send(jglobal.CMD_CLOSE, []byte{})
	kc.con.Close()
}

func (kc *Kcp) send(cmd uint16, data []byte) {
	pack := &jkcp.Pack{
		Cmd:  cmd,
		Data: data,
	}
	bodySize := gCmdSize + len(pack.Data)
	size := gHeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[gHeadSize:], pack.Cmd)
	copy(buffer[gHeadSize+gCmdSize:], pack.Data)
	for pos := 0; pos < size; {
		n, err := kc.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (kc *Kcp) recv() {
	buffer := make([]byte, gHeadSize)
	io.ReadFull(kc.con, buffer)
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	io.ReadFull(kc.con, buffer)
	pack := jkcp.Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
	jlog.Info(jdebug.StructToString(pack))
}
