package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jlog "jamger/log"
	jtcp "jamger/net/tcp"
	"net"
	"strings"
)

func testTcp() {
	jlog.Info("<test tcp>")
	addr := strings.Split(jconfig.GetString("tcp.addr"), ":")
	con, err := net.Dial("tcp", "127.0.0.1:"+addr[1])
	if err != nil {
		jlog.Fatal(err)
	}
	defer con.Close()
	jlog.Info("connect to server ", addr)

	pack := &jtcp.Pack{
		Cmd:  2,
		Data: []byte("hello world"),
	}
	sendTcpPack(con, pack)

	pack = recvTcpPack(con)
	jlog.Infoln(pack.Cmd, string(pack.Data))
}

func sendTcpPack(con net.Conn, pack *jtcp.Pack) {
	bodySize := gCmdSize + len(pack.Data)
	size := gHeadSize + bodySize
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, uint16(bodySize))
	binary.LittleEndian.PutUint16(buffer[gHeadSize:], pack.Cmd)
	copy(buffer[gHeadSize+gCmdSize:], pack.Data)
	for pos := 0; pos < size; {
		n, err := con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func recvTcpPack(con net.Conn) *jtcp.Pack {
	buffer := make([]byte, gHeadSize)
	if _, err := io.ReadFull(con, buffer); err != nil {
		jlog.Fatal(err)
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	if _, err := io.ReadFull(con, buffer); err != nil {
		jlog.Fatal(err)
	}
	return &jtcp.Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
}
