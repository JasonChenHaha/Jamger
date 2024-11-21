package main

import (
	"encoding/binary"
	"io"
	jconfig "jamger/config"
	jglobal "jamger/global"
	jlog "jamger/log"
	jkcp "jamger/net/kcp"
	"strings"

	"github.com/xtaci/kcp-go"
)

func testKcp() {
	jlog.Info("<test kcp>")
	addr := strings.Split(jconfig.GetString("kcp.addr"), ":")
	con, err := kcp.DialWithOptions("127.0.0.1:"+addr[1], nil, jconfig.GetInt("kcp.dataShards"), jconfig.GetInt("kcp.parityShards"))
	if err != nil {
		jlog.Fatal(err)
	}
	defer closeKcp(con)

	pack := &jkcp.Pack{
		Cmd:  2,
		Data: []byte("this is kcp"),
	}
	sendKcpPack(con, pack)

	select {}

	// pack = recvKcpPack(con)
	// jlog.Infoln(pack.Cmd, string(pack.Data))
}

func closeKcp(con *kcp.UDPSession) {
	pack := &jkcp.Pack{
		Cmd: jglobal.CMD_CLOSE,
	}
	sendKcpPack(con, pack)
	con.Close()
}

func sendKcpPack(con *kcp.UDPSession, pack *jkcp.Pack) {
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

func recvKcpPack(con *kcp.UDPSession) *jkcp.Pack {
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
