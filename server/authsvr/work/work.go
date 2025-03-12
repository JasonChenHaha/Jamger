package jwork

import (
	"jglobal"
	"jnet"
	"jpb"
	"jrpc"
	"juser"

	"go.mongodb.org/mongo-driver/mongo"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_AUTH)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_SIGN_UP_REQ, signUp, &jpb.SignUpReq{})
	jnet.Rpc.Register(jpb.CMD_SIGN_IN_REQ, signIn, &jpb.SignInReq{})
}

// ------------------------- inside.method -------------------------

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
	if err := juser.IsUserExist(req.Id); err == nil {
		rsp.Code = jpb.CODE_USER_EXIST
		return
	} else if err != mongo.ErrNoDocuments {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	} else {
		// 密码加密
		secret, err := juser.EncryptPwd(req.Pwd)
		if err != nil {
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
		// 生成用户id
		uid, err := juser.GenUserUid()
		if err != nil {
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
		// 创建
		if err = juser.CreateUser(uid, req.Id, secret); err != nil {
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
	if uid, err := juser.AccountCheck(req.Id, req.Pwd); err != nil {
		rsp.Code = jpb.CODE_ACCOUNT_FAIL
	} else {
		rsp.Uid = uid
	}
}
