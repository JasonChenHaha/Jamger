package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"hash/crc32"
	"io"
	"jconfig"
	"jglobal"
	"jlog"
	"jpb"
	"net/http"

	"google.golang.org/protobuf/proto"
)

const (
	checksumSize = 4
	aesKeySize   = 16
	cmdSize      = 2
)

type Http struct {
	pubKey *rsa.PublicKey
	aesKey []byte
	rsp    proto.Message
}

func testHttp() {
	jlog.Info("<test rpc>")
	htp := &Http{}
	pubKey, err := jglobal.RSALoadPublicKey(jconfig.GetString("rsa.publicKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	htp.pubKey = pubKey

	htp.aesKey, err = jglobal.AESGenerate(16)
	if err != nil {
		jlog.Fatal(err)
	}
	// htp.rsp = &jpb.Pong{}
	// msg := htp.encode(jpb.CMD_PING, &jpb.Ping{})
	// htp.send(msg)
	htp.rsp = &jpb.SignUpRsp{}
	msg := htp.encode(jpb.CMD_SIGN_UP_REQ, &jpb.SignUpReq{
		Id:  "heihei",
		Pwd: "123456",
	})
	htp.send(msg)
}

func (htp *Http) send(raw []byte) {
	addr := jconfig.GetString("http.addr")
	rsp, err := http.Post("http://"+addr, "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	cmd, msg := htp.decode(body)
	jlog.Infoln(cmd, msg)
}

func (htp *Http) encode(cmd jpb.CMD, msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Fatal(err)
	}
	dataSize := len(data)
	raw := make([]byte, cmdSize+dataSize+aesKeySize+checksumSize)
	binary.LittleEndian.PutUint16(raw, uint16(cmd))
	copy(raw[cmdSize:], data)
	copy(raw[cmdSize+dataSize:], htp.aesKey)
	binary.LittleEndian.PutUint32(raw[cmdSize+dataSize+aesKeySize:], crc32.ChecksumIEEE(raw[:cmdSize+dataSize+aesKeySize]))
	err = jglobal.RSAEncrypt(htp.pubKey, &raw)
	if err != nil {
		jlog.Fatal(err)
	}
	return raw
}

func (htp *Http) decode(raw []byte) (jpb.CMD, proto.Message) {
	err := jglobal.AESDecrypt(htp.aesKey, &raw)
	if err != nil {
		jlog.Fatal(err)
	}
	cmd := jpb.CMD(binary.LittleEndian.Uint16(raw))
	err = proto.Unmarshal(raw[cmdSize:], htp.rsp)
	if err != nil {
		jlog.Fatal(err)
	}
	return cmd, htp.rsp
}
