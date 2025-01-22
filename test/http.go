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
	cmdSize      = 2
	uidSize      = 4
	aesKeySize   = 16
	checksumSize = 4
)

type Http struct {
	pubKey *rsa.PublicKey
	aesKey []byte
	uid    uint32
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

	// htp.rsp = &jpb.SignUpRsp{}
	// cmd, msg := htp.sendAuth(jpb.CMD_SIGN_UP_REQ, &jpb.SignUpReq{
	// 	Id:  "gege",
	// 	Pwd: "123456",
	// })

	// htp.rsp = &jpb.SignInRsp{}
	// cmd, msg := htp.sendAuth(jpb.CMD_SIGN_IN_REQ, &jpb.SignInReq{
	// 	Id:  "gaga",
	// 	Pwd: "123456",
	// })

	htp.rsp = &jpb.Pong{}
	cmd, msg := htp.send(jpb.CMD_PING, &jpb.Ping{})

	jlog.Debugf("cmd: %s, msg: %s", cmd, msg)
}

func (htp *Http) sendAuth(cmd jpb.CMD, msg proto.Message) (jpb.CMD, proto.Message) {
	raw := htp.encodeRSA(cmd, msg)
	addr := jconfig.GetString("http.addr")
	rsp, err := http.Post("http://"+addr+"/auth", "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	return htp.decode(body)
}

func (htp *Http) send(cmd jpb.CMD, msg proto.Message) (jpb.CMD, proto.Message) {
	raw := htp.encodeAES(cmd, msg)
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
	return htp.decode(body)
}

func (htp *Http) encodeRSA(cmd jpb.CMD, msg proto.Message) []byte {
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

func (htp *Http) encodeAES(cmd jpb.CMD, msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Fatal(err)
	}
	dataSize := len(data)
	raw := make([]byte, cmdSize+dataSize+checksumSize)
	binary.LittleEndian.PutUint16(raw, uint16(cmd))
	copy(raw[cmdSize:], data)
	binary.LittleEndian.PutUint32(raw[cmdSize+dataSize:], crc32.ChecksumIEEE(raw[:cmdSize+dataSize]))
	if err = jglobal.AESEncrypt(htp.aesKey, &raw); err != nil {
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
	if cmd == jpb.CMD_GATE_INFO {
		htp.rsp = &jpb.Error{}
	}
	err = proto.Unmarshal(raw[cmdSize:], htp.rsp)
	if err != nil {
		jlog.Fatal(err)
	}
	return cmd, htp.rsp
}
