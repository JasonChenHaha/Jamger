package jwork

import (
	"jglobal"
	"jimage"
	"jnet"
	"jpb"
	"jrpc"
	"jschedule"
	"juser"
	"time"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_CENTER)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DelUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Register(jpb.CMD_GOOD_LIST_REQ, goodList, &jpb.GoodListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_GOOD_REQ, uploadGood, &jpb.UploadGoodReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_GOOD_REQ, modifyGood, &jpb.ModifyGoodReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_GOOD_REQ, deleteGood, &jpb.DeleteGoodReq{})
	jnet.Rpc.Register(jpb.CMD_IMAGE_REQ, image, &jpb.ImageReq{})
}

// ------------------------- inside.method -------------------------

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
		rsp.Goods = append(rsp.Goods, v)
	}
}

// 上传商品
func uploadGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	req := pack.Data.(*jpb.UploadGoodReq)
	rsp := &jpb.UploadGoodRsp{}
	pack.Cmd = jpb.CMD_UPLOAD_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	// 图片压缩
	image, err := jimage.Image.Compress(req.Good.Image)
	if err != nil {
		rsp.Code = jpb.CODE_IMAGE_ERR
		return
	}
	req.Good.Image = image
	// 生成商品uid
	user0 := juser.GetUserAnyway(0)
	uid, err := user0.GenGoodUid()
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	// 保存图片
	if err = jimage.Image.Add(uid, req.Good.Image); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	req.Good.Image = nil
	user0.AddGood(uid, req.Good)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DelUserReq{Uid: 0},
		})
	})
}

// 修改商品
func modifyGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	req := pack.Data.(*jpb.ModifyGoodReq)
	rsp := &jpb.ModifyGoodRsp{}
	pack.Cmd = jpb.CMD_MODIFY_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	// 图片压缩
	image, err := jimage.Image.Compress(req.Good.Image)
	if err != nil {
		rsp.Code = jpb.CODE_IMAGE_ERR
		return
	}
	req.Good.Image = image
	// 保存图片
	if err = jimage.Image.Add(req.Good.Uid, req.Good.Image); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	user0 := juser.GetUserAnyway(0)
	user0.ModifyGood(req.Good)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DelUserReq{Uid: 0},
		})
	})
}

// 下架商品
func deleteGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	req := pack.Data.(*jpb.DeleteGoodReq)
	rsp := &jpb.DeleteGoodRsp{}
	pack.Cmd = jpb.CMD_DELETE_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user0 := juser.GetUserAnyway(0)
	user0.DelGood(req.Uid)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DelUserReq{Uid: 0},
		})
	})
}

// 下载图片
func image(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.ImageReq)
	rsp := &jpb.ImageRsp{}
	pack.Cmd = jpb.CMD_IMAGE_RSP
	pack.Data = rsp
	image, err := jimage.Image.Get(req.Uid)
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
	} else {
		rsp.Image = image
	}
}
