package jwork

import (
	"jglobal"
	"jnet"
	"jpb"
	"jrpc"
	"juser"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_CENTER)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DelUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Register(jpb.CMD_GOOD_LIST_REQ, goodList, &jpb.GooDListReq{})
}

// ------------------------- inside -------------------------

// 缓存清理
func deleteUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.DelUserReq)
	pack.Data = &jpb.DelUserRsp{}
	if user := juser.GetUser(req.Uid); user != nil {
		user.Destory()
	}
}

// 登录
func login(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	rsp := &jpb.LoginRsp{}
	pack.Cmd = jpb.CMD_LOGIN_RSP
	pack.Data = rsp
	user.SetLoginTs()
}

// 获取商品列表
func goodList(pack *jglobal.Pack) {
	rsp := &jpb.GoodListRsp{Goods: []*jpb.Good{}}
	pack.Cmd = jpb.CMD_GOOD_LIST_RSP
	pack.Data = rsp
	user := juser.GetUserAnyway(0)
	for _, v := range user.Goods.Data {
		good := &jpb.Good{
			Id:      v.Id,
			Name:    v.Name,
			Desc:    v.Desc,
			Size:    v.Size,
			Price:   v.Price,
			ImgType: v.ImgType,
			Image:   v.Image,
		}
		rsp.Goods = append(rsp.Goods, good)
	}
}
