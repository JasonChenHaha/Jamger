package jwork

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jnet"
	"jpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

// ------------------------- outside -------------------------

func Init() {
	jnet.Tcp.Register(jpb.CMD_SIGN_UP_REQ, signUp, &jpb.SignUpReq{})
	jnet.Tcp.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
}

// ------------------------- inside -------------------------

// 注册
func signUp(id uint64, cmd jpb.CMD, msg proto.Message) {
	req := msg.(*jpb.SignUpReq)
	rsp := &jpb.SignUpRsp{}
	defer jnet.Tcp.Send(id, jpb.CMD_SIGN_UP_RSP, rsp)
	// 格式校验
	if len(req.Id) == 0 || len(req.Pwd) == 0 {
		rsp.Code = jpb.CODE_ACCOUNT_SYNTX
		return
	}
	// 判断账号是否存在
	in := &jmongo.Input{
		Col:    jglobal.MONGO_ACCOUNT,
		Filter: bson.M{"id": req.Id},
	}
	if err := jdb.Mongo.FindOne(in, &bson.M{}); err == nil {
		rsp.Code = jpb.CODE_ACCOUNT_EXIST
	} else if err != mongo.ErrNoDocuments {
		rsp.Code = jpb.CODE_SVR_ERR
	} else {
		// 创建账号
		secret, err := bcrypt.GenerateFromPassword([]byte(req.Pwd), bcrypt.DefaultCost)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
		}
		in := &jmongo.Input{
			Col:    jglobal.MONGO_ACCOUNT,
			Insert: bson.M{"id": req.Id, "pwd": secret},
		}
		err = jdb.Mongo.InsertOne(in)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
		}
	}
}

// 登录
func signIn(id uint64, cmd jpb.CMD, msg proto.Message) {
	req := msg.(*jpb.SignInReq)
	rsp := &jpb.SignInRsp{}
	defer jnet.Tcp.Send(id, jpb.CMD_SIGN_IN_RSP, rsp)
	// 格式校验
	if len(req.Id) == 0 || len(req.Pwd) == 0 {
		rsp.Code = jpb.CODE_ACCOUNT_SYNTX
		return
	}
	// 账号校验
	in := &jmongo.Input{
		Col:    jglobal.MONGO_ACCOUNT,
		Filter: bson.M{"id": req.Id},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOne(in, &out); err == mongo.ErrNoDocuments {
		// 账号不存在
		rsp.Code = jpb.CODE_ACCOUNT_FAIL
		return
	} else if err != nil {
		jlog.Error(err)
		rsp.Code = jpb.CODE_SVR_ERR
		return
	} else {
		secret := out["pwd"].(primitive.Binary)
		err := bcrypt.CompareHashAndPassword(secret.Data, []byte(req.Pwd))
		if err != nil {
			// 密码错误
			rsp.Code = jpb.CODE_ACCOUNT_FAIL
			return
		}
	}
	// 校验通过
	token, err := jglobal.TokenGenerate(req.Id)
	if err != nil {
		jlog.Error(err)
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	// 缓存
	if _, err = jdb.Redis.Do("SET", req.Id, token); err != nil {
		jlog.Error(err)
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	rsp.Token = token
}
