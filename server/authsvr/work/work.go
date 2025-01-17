package jwork

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"jnet"
	"jpb"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

// ------------------------- outside -------------------------

func Init() {
	jnet.Http.Register(jpb.CMD_SIGN_UP_REQ, signUp, &jpb.SignUpReq{})
	jnet.Http.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
}

// ------------------------- inside -------------------------

// 注册
func signUp(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message) {
	jlog.Debug("ok")
	req := msg.(*jpb.SignUpReq)
	rsp := &jpb.SignUpRsp{}
	defer jnet.Http.Response(w, jpb.CMD_SIGN_UP_RSP, rsp)
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
		// 获取自增id
		in := &jmongo.Input{
			Col:    jglobal.MONGO_ACCOUNT,
			Filter: bson.M{"_id": 0},
			Update: bson.M{"$inc": bson.M{"counter": 1}},
			Upsert: true,
			RetDoc: options.After,
		}
		out := bson.M{}
		err = jdb.Mongo.FindOneAndUpdate(in, &out)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
		}
		// 创建
		in = &jmongo.Input{
			Col:    jglobal.MONGO_ACCOUNT,
			Insert: bson.M{"_id": out["counter"], "id": req.Id, "pwd": secret},
		}
		err = jdb.Mongo.InsertOne(in)
		if err != nil {
			jlog.Error(err)
			rsp.Code = jpb.CODE_SVR_ERR
		}
	}
}

// 登录
func signIn(w http.ResponseWriter, cmd jpb.CMD, msg proto.Message) {
	req := msg.(*jpb.SignInReq)
	rsp := &jpb.SignInRsp{}
	defer jnet.Http.Response(w, jpb.CMD_SIGN_IN_RSP, rsp)
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
