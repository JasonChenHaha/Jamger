package jwork

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jnet"
	"jpb"
	"jrpc"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_AUTH)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_SIGN_UP_REQ, signUp, &jpb.SignUpReq{})
	jnet.Rpc.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
}

// ------------------------- inside -------------------------

// 注册
func signUp(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.SignUpReq)
	rsp := &jpb.SignUpRsp{}
	pack.Cmd = jpb.CMD_SIGN_UP_RSP
	pack.Data = rsp
	// 格式校验
	if len(req.Id) == 0 || len(req.Pwd) == 0 {
		rsp.Code = jpb.CODE_ACCOUNT_SYNTX
		return
	}
	// 判断账号是否存在
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"basic.id": req.Id},
		Project: bson.M{"_id": 1},
	}
	if err := jdb.Mongo.FindOne(in, &bson.M{}); err == nil {
		rsp.Code = jpb.CODE_USER_EXIST
		return
	} else if err != mongo.ErrNoDocuments {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	} else {
		// 创建账号
		secret, err := bcrypt.GenerateFromPassword([]byte(req.Pwd), bcrypt.DefaultCost)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
		// 获取自增id
		in := &jmongo.Input{
			Col:     jglobal.MONGO_USER,
			Filter:  bson.M{"_id": int64(0)},
			Update:  bson.M{"$inc": bson.M{"counter": int64(1)}},
			Upsert:  true,
			RetDoc:  options.After,
			Project: bson.M{"counter": 1},
		}
		out := bson.M{}
		err = jdb.Mongo.FindOneAndUpdate(in, &out)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
		// 创建
		in = &jmongo.Input{
			Col:    jglobal.MONGO_USER,
			Insert: bson.M{"_id": out["counter"], "basic": bson.M{"id": req.Id, "pwd": secret}},
		}
		err = jdb.Mongo.InsertOne(in)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
	}
}

// 登录
func signIn(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.SignInReq)
	rsp := &jpb.SignInRsp{}
	pack.Cmd = jpb.CMD_SIGN_IN_RSP
	pack.Data = rsp
	// 格式校验
	if len(req.Id) == 0 || len(req.Pwd) == 0 {
		rsp.Code = jpb.CODE_ACCOUNT_SYNTX
		return
	}
	// 账号校验
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"basic.id": req.Id},
		Project: bson.M{"basic.pwd": 1},
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
		secret := out["basic"].(primitive.M)["pwd"].(primitive.Binary)
		err := bcrypt.CompareHashAndPassword(secret.Data, []byte(req.Pwd))
		if err != nil {
			// 密码错误
			rsp.Code = jpb.CODE_ACCOUNT_FAIL
			return
		}
	}
	// 校验通过
	rsp.Uid = uint32(out["_id"].(int64))
}
