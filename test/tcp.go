package main

import (
	"encoding/binary"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type Tcp struct {
	con    net.Conn
	aesKey []byte
	rsp    proto.Message
}

func testTcp() {
	jlog.Info("<test tcp>")
	tcp := &Tcp{}
	addr := jconfig.GetString("tcp.addr")
	con, _ := net.Dial("tcp", addr)
	jlog.Info("connect to server ", addr)
	tcp.con = con

	// tcp.rsp = &jpb.SignInRsp{}
	// tcp.sendWithRsa(jpb.CMD_SIGN_IN_REQ, &jpb.SignInReq{
	// 	Id:  "nihao",
	// 	Pwd: "123456",
	// })
	// tcp.recv()

	// tcp.sendWithAes(jpb.CMD_PING, nil)
	// tcp.recv()
	// go tcp.heartbeat()
	// jglobal.Keep()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		tcp.sendWithAes(jpb.CMD_HEARTBEAT, nil)
	}
}

func (tcp *Tcp) sendWithAes(cmd jpb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	jglobal.AesEncrypt(tcp.aesKey, &data)
	size := CmdSize + len(data)
	buffer := make([]byte, HeadSize+size)
	binary.LittleEndian.PutUint16(buffer, uint16(size))
	binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(cmd))
	copy(buffer[HeadSize+CmdSize:], data)
	for pos := 0; pos < size; {
		n, err := tcp.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) recv() {
	buffer := make([]byte, HeadSize)
	_, err := io.ReadFull(tcp.con, buffer)
	if err == io.EOF {
		jlog.Debug("close by server")
		return
	}
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	io.ReadFull(tcp.con, buffer)
	cmd := jpb.CMD(binary.LittleEndian.Uint16(buffer))
	data := buffer[CmdSize:]
	jglobal.AesDecrypt(tcp.aesKey, &data)
	proto.Unmarshal(data, tcp.rsp)
	jlog.Infoln(cmd, tcp.rsp)
}
