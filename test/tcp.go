package main

import (
	"crypto/rsa"
	"encoding/binary"
	"hash/crc32"
	"io"
	"jconfig"
	"jdebug"
	"jglobal"
	"jlog"
	pb "jpb"
	"jtcp"
	"jtrash"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

type Tcp struct {
	con    net.Conn
	pubKey *rsa.PublicKey
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

	// go tcp.heartbeat()

	tcp.send(jglobal.CMD_SIGN_UP_REQ, &pb.SignUpReq{
		Id:  "nihaoya",
		Pwd: "123456",
	})
	// tcp.recv()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		tcp.send(jglobal.CMD_HEARTBEAT, nil)
	}
}

func (tcp *Tcp) send(cmd uint16, msg proto.Message) {
	data := []byte{}
	if msg != nil {
		data, _ = proto.Marshal(msg)
	}
	pack := &jtcp.Pack{
		Cmd:  cmd,
		Data: data,
	}
	size := len(pack.Data)
	raw := make([]byte, size+4)
	copy(raw, pack.Data)
	binary.LittleEndian.PutUint32(raw[size:], crc32.ChecksumIEEE(pack.Data))
	secret, err := jtrash.RSAEncrypt(tcp.pubKey, raw)
	if err != nil {
		jlog.Fatal(err)
	}
	size = gCmdSize + len(secret)
	buffer := make([]byte, gHeadSize+size)
	binary.LittleEndian.PutUint16(buffer, uint16(size))
	binary.LittleEndian.PutUint16(buffer[gHeadSize:], pack.Cmd)
	copy(buffer[gHeadSize+gCmdSize:], secret)
	for pos := 0; pos < size; {
		n, err := tcp.con.Write(buffer)
		if err != nil {
			jlog.Fatal(err)
		}
		pos += n
	}
}

func (tcp *Tcp) recv() {
	buffer := make([]byte, gHeadSize)
	io.ReadFull(tcp.con, buffer)
	bodySize := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, bodySize)
	io.ReadFull(tcp.con, buffer)
	pack := &jtcp.Pack{
		Cmd:  binary.LittleEndian.Uint16(buffer),
		Data: buffer[gCmdSize:],
	}
	jlog.Info(jdebug.StructToString(pack))
}
