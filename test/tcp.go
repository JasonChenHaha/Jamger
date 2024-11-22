package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jdebug "jamger/debug"
	jglobal "jamger/global"
	jlog "jamger/log"
	jtcp "jamger/net/tcp"
	"net"
	"strings"
	"time"
)

type Tcp struct {
	con net.Conn
}

func testTcp() *Tcp {
	jlog.Info("<test tcp>")
	tcp := &Tcp{}
	addr := strings.Split(jconfig.GetString("tcp.addr"), ":")
	con, _ := net.Dial("tcp", "127.0.0.1:"+addr[1])
	jlog.Info("connect to server ", addr)
	tcp.con = con

	go tcp.heartbeat()

	tcp.send(jglobal.CMD_PING, []byte{})
	tcp.recv()

	return tcp
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		tcp.send(jglobal.CMD_HEARTBEAT, []byte{})
	}
}

func (tcp *Tcp) send(cmd uint16, data []byte) {
	pack := &jtcp.Pack{
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
		n, err := tcp.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) recv() {
	buffer := make([]byte, gHeadSize)
	io.ReadFull(tcp.con, buffer)
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	io.ReadFull(tcp.con, buffer)
	pack := &jtcp.Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
	jlog.Info(jdebug.StructToString(pack))
}
