package main

import (
	"encoding/binary"
	"io"
	"jconfig"
	"jdebug"
	"jkcp"
	"jlog"
	pb "jpb"
	"time"

	"github.com/xtaci/kcp-go"
)

type Kcp struct {
	con *kcp.UDPSession
}

func testKcp() {
	jlog.Info("<test kcp>")
	kc := &Kcp{}
	addr := jconfig.GetString("kcp.addr")
	con, _ := kcp.DialWithOptions(addr, nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	jlog.Info("connect to server ", addr)
	kc.con = con

	go kc.heartbeat()

	kc.send(pb.CMD_PING, []byte{})
	kc.recv()
}

func (kc *Kcp) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		kc.send(pb.CMD_HEARTBEAT, []byte{})
	}
}

func (kc *Kcp) close() {
	kc.send(pb.CMD_CLOSE, []byte{})
	kc.con.Close()
}

func (kc *Kcp) send(cmd pb.CMD, data []byte) {
	pack := &jkcp.Pack{
		Cmd:  cmd,
		Data: data,
	}
	bodySize := CmdSize + len(pack.Data)
	size := HeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(pack.Cmd))
	copy(buffer[HeadSize+CmdSize:], pack.Data)
	for pos := 0; pos < size; {
		n, err := kc.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (kc *Kcp) recv() {
	buffer := make([]byte, HeadSize)
	io.ReadFull(kc.con, buffer)
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	io.ReadFull(kc.con, buffer)
	pack := jkcp.Pack{
		Cmd:  pb.CMD(binary.LittleEndian.Uint16(buffer)),
		Data: buffer[CmdSize:],
	}
	jlog.Info(jdebug.StructToString(pack))
}
