package jwork

import (
	"jglobal"
	"jlog"
	"jnet"
	pb "jpb"

	"google.golang.org/protobuf/proto"
)

func Init() {
	jnet.Tcp.Register(jglobal.CMD_SIGN_UP_REQ, signUp, &pb.SignUpReq{})
}

func signUp(id uint64, cmd uint16, msg proto.Message) {
	req := msg.(*pb.SignUpReq)
	jlog.Debug(req)
	rsp := &pb.SignUpRsp{Code: jglobal.CMD_OK}
	jnet.Tcp.Send(id, jglobal.CMD_SIGN_UP_RSP, rsp)
}
