package main

import (
	"encoding/binary"
	"hash/crc32"
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
	con net.Conn
	cmd jpb.CMD
	rsp proto.Message
}

func testTcp() {
	jlog.Info("<test tcp>")
	tcp := &Tcp{}
	addr := jconfig.GetString("tcp.addr")
	con, _ := net.Dial("tcp", addr)
	jlog.Info("connect to server ", addr)
	tcp.con = con

	tcp.rsp = &jpb.LoginRsp{}
	tcp.sendWithAes(jpb.CMD_LOGIN_REQ, &jpb.LoginReq{})
	tcp.recv()

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
	size := len(data)
	raw := make([]byte, cmdSize+size+checksumSize)
	binary.LittleEndian.PutUint16(raw, uint16(cmd))
	copy(raw[cmdSize:], data)
	binary.LittleEndian.PutUint32(raw[cmdSize+size:], crc32.ChecksumIEEE(raw[:cmdSize+size]))
	jglobal.AesEncrypt(aesKey, &raw)
	size = len(raw)
	raw2 := make([]byte, packSize+uidSize+size)
	binary.LittleEndian.PutUint16(raw2, uint16(uidSize+size))
	binary.LittleEndian.PutUint32(raw2[packSize:], uid)
	copy(raw2[packSize+uidSize:], raw)
	for pos := 0; pos < size; {
		n, err := tcp.con.Write(raw2)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) recv() {
	raw := make([]byte, packSize)
	_, err := io.ReadFull(tcp.con, raw)
	if err == io.EOF {
		jlog.Debug("close by server")
		return
	}
	size := binary.LittleEndian.Uint16(raw)
	raw = make([]byte, size)
	io.ReadFull(tcp.con, raw)
	jglobal.AesDecrypt(aesKey, &raw)
	tcp.cmd = jpb.CMD(binary.LittleEndian.Uint16(raw))
	proto.Unmarshal(raw[cmdSize:], tcp.rsp)
	jlog.Infof("cmd = %d, msg = %s", tcp.cmd, tcp.rsp)
}
