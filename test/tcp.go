package main

import (
	"crypto/rsa"
	"encoding/binary"
	"hash/crc32"
	"io"
	"jconfig"
	"jlog"
	pb "jpb"
	"jtrash"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type Tcp struct {
	con    net.Conn
	pubKey *rsa.PublicKey
	aesKey []byte
	msg    proto.Message
}

func testTcp() {
	jlog.Info("<test tcp>")
	tcp := &Tcp{}
	addr := jconfig.GetString("tcp.addr")
	con, _ := net.Dial("tcp", addr)
	jlog.Info("connect to server ", addr)
	tcp.con = con

	pubKey, err := jtrash.RSALoadPublicKey(jconfig.GetString("rsa.publicKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	tcp.pubKey = pubKey

	tcp.aesKey, err = jtrash.AESGenerate(AesKeySize)
	if err != nil {
		jlog.Fatal(err)
	}

	tcp.msg = &pb.SignUpRsp{}
	tcp.sendWithRSA(pb.CMD_SIGN_UP_REQ, &pb.SignUpReq{
		Id:  "nihao",
		Pwd: "123456",
	})
	tcp.recv()
	tcp.sendWithAES(pb.CMD_PING, nil)
	tcp.recv()
	// go tcp.heartbeat()
	jtrash.Keep()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		tcp.sendWithAES(pb.CMD_HEARTBEAT, nil)
	}
}

func (tcp *Tcp) sendWithRSA(cmd pb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	size := len(data)
	raw := make([]byte, size+AesKeySize+ChecksumSize)
	copy(raw, data)
	copy(raw[size:], tcp.aesKey)
	binary.LittleEndian.PutUint32(raw[size+AesKeySize:], crc32.ChecksumIEEE(raw[:size+AesKeySize]))
	if err := jtrash.RSAEncrypt(tcp.pubKey, &raw); err != nil {
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

func (tcp *Tcp) sendWithAES(cmd pb.CMD, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	size := len(data)
	data = append(data, make([]byte, 4)...)
	binary.LittleEndian.PutUint32(data[size:], crc32.ChecksumIEEE(data[:size]))
	jtrash.AESEncrypt(tcp.aesKey, &data)
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
	cmd := pb.CMD(binary.LittleEndian.Uint16(buffer))
	data := buffer[CmdSize:]
	jtrash.AESDecrypt(tcp.aesKey, &data)
	proto.Unmarshal(data[:len(data)-4], tcp.msg)
	jlog.Infoln(cmd, tcp.msg)
}
