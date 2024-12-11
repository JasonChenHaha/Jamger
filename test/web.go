package main

import (
	"encoding/binary"
	"jconfig"
	"jdebug"
	"jglobal"
	"jlog"
	"jweb"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Web struct {
	con *websocket.Conn
}

func testWeb() *Web {
	jlog.Info("<test web>")
	web := &Web{}
	addr := strings.Split(jconfig.GetString("web.addr"), ":")
	con, _, _ := websocket.DefaultDialer.Dial("ws://127.0.0.1:"+addr[1]+"/ws", nil)
	jlog.Info("connect to server ", addr)
	web.con = con

	go web.heartbeat()

	web.send(jglobal.CMD_PING, []byte{})
	web.recv()

	return web
}

func (web *Web) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		web.send(jglobal.CMD_HEARTBEAT, []byte{})
	}
}

func (web *Web) send(cmd uint16, data []byte) {
	pack := &jweb.Pack{
		Cmd:  cmd,
		Data: data,
	}
	sData := web.serializePack(pack)
	web.con.WriteMessage(websocket.BinaryMessage, sData)
}

func (web *Web) recv() {
	_, data, _ := web.con.ReadMessage()
	pack := web.unserializeToPack(data)
	jlog.Info(jdebug.StructToString(pack))
}

func (web *Web) unserializeToPack(data []byte) *jweb.Pack {
	return &jweb.Pack{
		Cmd:  binary.LittleEndian.Uint16(data),
		Data: data[gCmdSize:],
	}
}

func (web *Web) serializePack(pack *jweb.Pack) []byte {
	size := gCmdSize + len(pack.Data)
	buffer := make([]byte, size)
	binary.LittleEndian.PutUint16(buffer, pack.Cmd)
	copy(buffer[gCmdSize:], pack.Data)
	return buffer
}
