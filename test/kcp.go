package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jglobal "jamger/global"
	jlog "jamger/log"
	jkcp "jamger/net/kcp"
	"net"
	"strings"

	"github.com/xtaci/kcp-go"
)

func testKcp() {
	jlog.Info("<test kcp>")
	addr := strings.Split(jconfig.Get("kcp.addr").(string), ":")
	con, err := kcp.DialWithOptions("127.0.0.1:"+addr[1], nil, jglobal.G_KCP_DATASHARDS, jglobal.G_KCP_PARITYSHARDS)
	if err != nil {
		jlog.Fatal(err)
	}
	defer con.Close()

	pack := &jkcp.Pack{
		Cmd:  2,
		Data: []byte("this is kcp"),
	}
	sendKcpPack(con, pack)

	pack = recvKcpPack(con)
	jlog.Infoln(pack.Cmd, string(pack.Data))
}

func sendKcpPack(con net.Conn, pack *jkcp.Pack) {
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

func recvKcpPack(con net.Conn) *jkcp.Pack {
	buffer := make([]byte, gHeadSize)
	if _, err := io.ReadFull(con, buffer); err != nil {
		jlog.Fatal(err)
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	if _, err := io.ReadFull(con, buffer); err != nil {
		jlog.Fatal(err)
	}
	return &jkcp.Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
}
