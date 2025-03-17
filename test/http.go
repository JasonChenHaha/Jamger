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

type Http struct {
	pubKey *rsa.PublicKey
	uid    uint32
	rsp    proto.Message
}

func testHttp() {
	jlog.Info("<test http>")
	htp := &Http{}
	pubKey, err := jglobal.RsaLoadPublicKey(jconfig.GetString("rsa.publicKey"))
	if err != nil {
		jlog.Fatal(err)
	}
	htp.pubKey = pubKey

	var cmd jpb.CMD
	var msg proto.Message

	htp.rsp = &jpb.SignUpRsp{}
	cmd, msg = htp.sendAuth(jpb.CMD_SIGN_UP_REQ, &jpb.SignUpReq{
		Id:  id,
		Pwd: pwd,
	})
	jlog.Debugf("cmd: %s, msg: %s", cmd, msg)

	htp.rsp = &jpb.SignInRsp{}
	cmd, msg = htp.sendAuth(jpb.CMD_SIGN_IN_REQ, &jpb.SignInReq{
		Id:  id,
		Pwd: pwd,
	})
	if cmd == jpb.CMD_GATE_INFO {
		jlog.Debugf("cmd: %s, msg: %s", cmd, msg)
		return
	} else {
		uid = msg.(*jpb.SignInRsp).Uid
		jlog.Debugf("cmd: %s, msg: %s", cmd, msg)
	}
}

func (htp *Http) sendAuth(cmd jpb.CMD, msg proto.Message) (jpb.CMD, proto.Message) {
	raw := htp.encodeRsa(cmd, msg)
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
	if len(body) == 0 {
		jlog.Fatal("body is empty")
	}
	return htp.decode(body)
}

func (htp *Http) send(cmd jpb.CMD, msg proto.Message) (jpb.CMD, proto.Message) {
	raw := htp.encodeAes(cmd, msg)
	rsp, err := http.Post("http://"+httpAddr, "", bytes.NewBuffer(raw))
	if err != nil {
		jlog.Fatal(err)
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		jlog.Fatal(err)
	}
	if len(body) == 0 {
		jlog.Fatal("body is empty")
	}
	return htp.decode(body)
}

func (htp *Http) encodeRsa(cmd jpb.CMD, msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Fatal(err)
	}
	dataSize := len(data)
	raw := make([]byte, dataSize+aesKeySize+checksumSize)
	copy(raw, data)
	copy(raw[dataSize:], aesKey)
	binary.LittleEndian.PutUint32(raw[dataSize+aesKeySize:], crc32.ChecksumIEEE(raw[:dataSize+aesKeySize]))
	err = jglobal.RsaEncrypt(htp.pubKey, &raw)
	if err != nil {
		jlog.Fatal(err)
	}
	dataSize = len(raw)
	raw2 := make([]byte, cmdSize+dataSize)
	binary.LittleEndian.PutUint16(raw2, uint16(cmd))
	copy(raw2[cmdSize:], raw)
	return raw2
}

func (htp *Http) encodeAes(cmd jpb.CMD, msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		jlog.Fatal(err)
	}
	dataSize := len(data)
	raw := make([]byte, dataSize+checksumSize)
	copy(raw, data)
	binary.LittleEndian.PutUint32(raw[dataSize:], crc32.ChecksumIEEE(raw[:dataSize]))
	if err = jglobal.AesEncrypt(aesKey, &raw); err != nil {
		jlog.Fatal(err)
	}
	raw2 := make([]byte, uidSize+cmdSize+len(raw))
	binary.LittleEndian.PutUint16(raw2, uint16(cmd))
	binary.LittleEndian.PutUint32(raw2[uidSize:], uint32(htp.uid))
	copy(raw2[uidSize+cmdSize:], raw)
	return raw2
}

func (htp *Http) decode(raw []byte) (jpb.CMD, proto.Message) {
	err := jglobal.AesDecrypt(aesKey, &raw)
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
