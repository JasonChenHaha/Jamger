package main

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type Tcp struct {
	con net.Conn
	msg map[jpb.CMD]proto.Message
}

func testTcp() {
	jlog.Info("<test tcp>")
	tcp := &Tcp{
		msg: map[jpb.CMD]proto.Message{},
	}
	con, _ := net.Dial("tcp", tcpAddr)
	jlog.Info("connect to server ", tcpAddr)
	tcp.con = con

	tcp.msg[jpb.CMD_GATE_INFO] = &jpb.Error{}
	tcp.msg[jpb.CMD_NOTIFY] = &jpb.Notify{}
	tcp.msg[jpb.CMD_LOGIN_RSP] = &jpb.LoginRsp{}

	go tcp.recv()

	tcp.sendWithAes(jpb.CMD_LOGIN_REQ, &jpb.LoginReq{})

	// tcp.sendWithAes(jpb.CMD_PING, nil)
	// tcp.recv()
	go tcp.heartbeat()
	jglobal.Keep()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		tcp.sendWithAes(jpb.CMD_HEARTBEAT, &jpb.HeartbeatReq{})
	}
}

func (tcp *Tcp) sendWithAes(cmd jpb.CMD, msg proto.Message) {
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
	raw2 := make([]byte, packSize+uidSize+cmdSize+size)
	binary.LittleEndian.PutUint16(raw2, uint16(uidSize+cmdSize+size))
	binary.LittleEndian.PutUint32(raw2[packSize:], uid)
	binary.LittleEndian.PutUint16(raw2[packSize+uidSize:], uint16(cmd))
	copy(raw2[packSize+uidSize+cmdSize:], raw)
	for pos := 0; pos < size; {
		n, err := tcp.con.Write(raw2)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) recv() {
	for {
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
		cmd := jpb.CMD(binary.LittleEndian.Uint16(raw))
		proto.Unmarshal(raw[cmdSize:], tcp.msg[cmd])
		jlog.Infof("cmd(%v), %v", cmd, tcp.msg[cmd])
	}
}
