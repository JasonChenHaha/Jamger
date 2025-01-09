package jwork

import (
	"jdb"
	"jglobal"
	"jmongo"
	"jnet"
	pb "jpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

func Init() {
	jnet.Tcp.Register(pb.CMD_SIGN_UP_REQ, signUp, &pb.SignUpReq{})
}

func signUp(id uint64, cmd pb.CMD, msg proto.Message) {
	req := msg.(*pb.SignUpReq)
	rsp := &pb.SignUpRsp{}
	// 判断账号是否存在
	in := &jmongo.Input{
		Col:    jglobal.MONGO_ACCOUNT,
		Filter: bson.M{"id": req.Id},
	}
	if err := jdb.Mongo.FindOne(in, &bson.M{}); err == nil {
		rsp.Code = pb.CODE_ACCOUNT_EXIST
	} else if err != mongo.ErrNoDocuments {
		rsp.Code = pb.CODE_ERR
	} else {
	}
	jnet.Tcp.Send(id, pb.CMD_SIGN_UP_RSP, rsp)
}
