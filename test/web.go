package main

import (
	"jconfig"
	"jdebug"
	"jlog"
	"jpb"
	"jweb"
	"time"

	"github.com/gorilla/websocket"
)

type Web struct {
	con *websocket.Conn
}

func testWeb() {
	jlog.Info("<test web>")
	web := &Web{}
	addr := jconfig.GetString("web.addr")
	con, _, _ := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	jlog.Info("connect to server ", addr)
	web.con = con

	go web.heartbeat()

	web.send(jpb.CMD_PING, []byte{})
	web.recv()
}

func (web *Web) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		web.send(jpb.CMD_HEARTBEAT, []byte{})
	}
}

func (web *Web) send(cmd jpb.CMD, data []byte) {
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
	// return &jweb.Pack{
	// 	Cmd:  jpb.CMD(binary.LittleEndian.Uint16(data)),
	// 	Data: data[CmdSize:],
	// }
	return nil
}

func (web *Web) serializePack(pack *jweb.Pack) []byte {
	// size := CmdSize + len(pack.Data)
	// buffer := make([]byte, size)
	// binary.LittleEndian.PutUint16(buffer, uint16(pack.Cmd))
	// copy(buffer[CmdSize:], pack.Data)
	// return buffer
	return nil
}
