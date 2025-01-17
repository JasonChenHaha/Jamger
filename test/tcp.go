package main

import (
	"crypto/rsa"
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
	con    net.Conn
	pubKey *rsa.PublicKey
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
	pubKey, err := jglobal.RSALoadPublicKey(jconfig.GetString("rsa.publicKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	tcp.pubKey = pubKey

	tcp.aesKey, err = jglobal.AESGenerate(AesKeySize)
	if err != nil {
		jlog.Fatal(err)
	}

	tcp.rsp = &jpb.SignUpRsp{}
	tcp.sendWithRSA(jpb.CMD_SIGN_UP_REQ, &jpb.SignUpReq{
		Id:  "nihao",
		Pwd: "123456",
	})
	tcp.recv()

	// tcp.rsp = &jpb.SignInRsp{}
	// tcp.sendWithRSA(jpb.CMD_SIGN_IN_REQ, &jpb.SignInReq{
	// 	Id:  "nihao",
	// 	Pwd: "123456",
	// })
	// tcp.recv()

	// tcp.sendWithAES(jpb.CMD_PING, nil)
	// tcp.recv()
	// go tcp.heartbeat()
	// jglobal.Keep()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		tcp.sendWithAES(jpb.CMD_HEARTBEAT, nil)
	}
}

func (tcp *Tcp) sendWithRSA(cmd jpb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	size := len(data)
	raw := make([]byte, size+AesKeySize+ChecksumSize)
	copy(raw, data)
	copy(raw[size:], tcp.aesKey)
	binary.LittleEndian.PutUint32(raw[size+AesKeySize:], crc32.ChecksumIEEE(raw[:size+AesKeySize]))
	if err := jglobal.RSAEncrypt(tcp.pubKey, &raw); err != nil {
		jlog.Fatal(err)
	}
	size = CmdSize + len(raw)
	buffer := make([]byte, HeadSize+size)
	binary.LittleEndian.PutUint16(buffer, uint16(size))
	binary.LittleEndian.PutUint16(buffer[HeadSize:], uint16(cmd))
	copy(buffer[HeadSize+CmdSize:], raw)
	for pos := 0; pos < size; {
		n, err := tcp.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) sendWithAES(cmd jpb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	size := len(data)
	data = append(data, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(data[size:], crc32.ChecksumIEEE(data[:size]))
	jglobal.AESEncrypt(tcp.aesKey, &data)
	size = CmdSize + len(data)
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
	jglobal.AESDecrypt(tcp.aesKey, &data)
	proto.Unmarshal(data[:len(data)-4], tcp.rsp)
	jlog.Infoln(cmd, tcp.rsp)
}
