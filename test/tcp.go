package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"jglobal"
	"jlog"
	"jpb"
	"net"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
)

type Tcp struct {
	con net.Conn
	req proto.Message
	rsp map[jpb.CMD]proto.Message
}

func testTcp() {
	jlog.Info("<test tcp>")
	tcp := &Tcp{
		rsp: map[jpb.CMD]proto.Message{},
	}
	con, _ := net.Dial("tcp", tcpAddr)
	jlog.Info("connect to server ", tcpAddr)
	tcp.con = con

	tcp.rsp[jpb.CMD_GATE_INFO] = &jpb.Error{}
	tcp.rsp[jpb.CMD_LOGIN_RSP] = &jpb.LoginRsp{}
	tcp.rsp[jpb.CMD_SWIPER_LIST_RSP] = &jpb.SwiperListRsp{}
	tcp.rsp[jpb.CMD_UPLOAD_SWIPER_RSP] = &jpb.UploadSwiperRsp{}
	tcp.rsp[jpb.CMD_DELETE_SWIPER_RSP] = &jpb.DeleteSwiperRsp{}
	tcp.rsp[jpb.CMD_GOOD_LIST_RSP] = &jpb.GoodListRsp{}
	tcp.rsp[jpb.CMD_UPLOAD_GOOD_RSP] = &jpb.UploadGoodRsp{}
	tcp.rsp[jpb.CMD_DELETE_GOOD_RSP] = &jpb.DeleteGoodRsp{}
	tcp.rsp[jpb.CMD_IMAGE_RSP] = &jpb.ImageRsp{}

	go tcp.recv()
	go tcp.heartbeat()

	tcp.send(jpb.CMD_LOGIN_REQ, &jpb.LoginReq{})

	// image, err := os.ReadFile("../template/4.jpg")
	// if err != nil {
	// 	jlog.Error(err)
	// 	return
	// }

	tcp.send(jpb.CMD_SWIPER_LIST_REQ, &jpb.SwiperListReq{})
	// tcp.send(jpb.CMD_UPLOAD_SWIPER_REQ, &jpb.UploadSwiperReq{
	// 	Image: image,
	// })
	tcp.send(jpb.CMD_DELETE_SWIPER_REQ, &jpb.DeleteSwiperReq{
		Uid: 26,
	})

	// tcp.send(jpb.CMD_GOOD_LIST_REQ, &jpb.GoodListRsp{})
	// tcp.send(jpb.CMD_IMAGE_REQ, &jpb.ImageReq{
	// 	Name: "1.jpeg",
	// })

	// tcp.send(jpb.CMD_UPLOAD_GOOD_REQ, &jpb.UploadGoodReq{
	// 	Good: &jpb.Good{
	// 		Name:  "商品",
	// 		Desc:  "描述",
	// 		Size:  "37",
	// 		Price: 100,
	// 		Image: image,
	// 		Kind:  "类别6",
	// 	},
	// })
	// tcp.send(jpb.CMD_MODIFY_GOOD_REQ, &jpb.ModifyGoodReq{
	// 	Good: &jpb.Good{
	// 		Uid:   3,
	// 		Name:  "name2",
	// 		Desc:  "desc2",
	// 		Size:  1,
	// 		Price: 1,
	// 		Image: image,
	// 	},
	// })
	// tcp.send(jpb.CMD_DELETE_GOOD_REQ, &jpb.DeleteGoodReq{
	// 	Uid: 24,
	// })
	jglobal.Keep()
}

func (tcp *Tcp) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		tcp.send(jpb.CMD_HEARTBEAT, &jpb.HeartbeatReq{})
	}
}

func (tcp *Tcp) send(cmd jpb.CMD, msg proto.Message) {
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
	binary.LittleEndian.PutUint32(raw2, uint32(uidSize+cmdSize+size))
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
	tcp.req = msg
}

func (tcp *Tcp) recv() {
	for {
		raw := make([]byte, packSize)
		_, err := io.ReadFull(tcp.con, raw)
		if err == io.EOF {
			jlog.Debug("close by server")
			return
		}
		size := binary.LittleEndian.Uint32(raw)
		raw = make([]byte, size)
		io.ReadFull(tcp.con, raw)
		jglobal.AesDecrypt(aesKey, &raw)
		cmd := jpb.CMD(binary.LittleEndian.Uint16(raw))
		proto.Unmarshal(raw[cmdSize:], tcp.rsp[cmd])
		switch cmd {
		case jpb.CMD_IMAGE_RSP:
			req := tcp.req.(*jpb.ImageReq)
			rsp := tcp.rsp[cmd].(*jpb.ImageRsp)
			file, err := os.Create(fmt.Sprintf("./%d", req.Uid))
			if err != nil {
				jlog.Error(err)
				return
			}
			file.Write(rsp.Image)
			file.Close()
		default:
			jlog.Infof("cmd(%v), %v", cmd, tcp.rsp[cmd])
		}
	}
}
