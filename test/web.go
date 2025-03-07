package main

import (
	"encoding/binary"
	"hash/crc32"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type Web struct {
	con *websocket.Conn
	msg map[jpb.CMD]proto.Message
}

func testWeb() {
	jlog.Info("<test web>")
	web := &Web{
		msg: map[jpb.CMD]proto.Message{},
	}
	addr := jconfig.GetString("web.addr")
	con, _, _ := websocket.DefaultDialer.Dial("ws://"+addr+"/ws", nil)
	jlog.Info("connect to server ", addr)
	web.con = con

	web.msg[jpb.CMD_GATE_INFO] = &jpb.Error{}
	web.msg[jpb.CMD_LOGIN_RSP] = &jpb.LoginRsp{}
	web.msg[jpb.CMD_GOOD_LIST_RSP] = &jpb.GoodListRsp{}
	web.msg[jpb.CMD_UPLOAD_GOOD_RSP] = &jpb.UploadGoodRsp{}

	go web.recv()
	go web.heartbeat()

	web.sendWithAes(jpb.CMD_LOGIN_REQ, &jpb.LoginReq{})

	jglobal.Keep()
}

func (web *Web) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		web.sendWithAes(jpb.CMD_HEARTBEAT, &jpb.HeartbeatReq{})
	}
}

func (web *Web) sendWithAes(cmd jpb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	size := len(data)
	raw := make([]byte, size+checksumSize)
	copy(raw, data)
	binary.LittleEndian.PutUint32(raw[size:], crc32.ChecksumIEEE(raw[:size]))
	jglobal.AesEncrypt(aesKey, &raw)
	size = len(raw)
	raw2 := make([]byte, uidSize+cmdSize+size)
	binary.LittleEndian.PutUint32(raw2, uid)
	binary.LittleEndian.PutUint16(raw2[uidSize:], uint16(cmd))
	copy(raw2[uidSize+cmdSize:], raw)
	web.con.WriteMessage(websocket.BinaryMessage, raw2)
}

func (web *Web) recv() {
	for {
		_, raw, err := web.con.ReadMessage()
		if err != nil {
			jlog.Error(err)
			return
		}
		jglobal.AesDecrypt(aesKey, &raw)
		cmd := jpb.CMD(binary.LittleEndian.Uint16(raw))
		proto.Unmarshal(raw[cmdSize:], web.msg[cmd])
		jlog.Infof("cmd(%v), %v", cmd, web.msg[cmd])
	}
}
