package main

import (
	"encoding/binary"
	jconfig "jamger/config"
	jlog "jamger/log"
	jweb "jamger/net/web"
	"strings"

	"github.com/gorilla/websocket"
)

func testWeb() {
	jlog.Info("<test web>")
	addr := strings.Split(jconfig.GetString("web.addr"), ":")
	con, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:"+addr[1]+"/ws", nil)
	if err != nil {
		jlog.Fatal(err)
	}
	pack := jweb.Pack{
		Cmd:  1,
		Data: []byte("web hello world"),
	}
	data := serializePack(&pack)
	if err = con.WriteMessage(websocket.BinaryMessage, data); err != nil {
		jlog.Fatal(err)
	}
	select {}
}

func unserializeData(data []byte) *jweb.Pack {
	return &jweb.Pack{
		Cmd:  binary.LittleEndian.Uint16(data),
		Data: data[gCmdSize:],
	}
}

func serializePack(pack *jweb.Pack) []byte {
	size := gCmdSize + len(pack.Data)
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, pack.Cmd)
	copy(buffer[gCmdSize:], pack.Data)
	return buffer
}
