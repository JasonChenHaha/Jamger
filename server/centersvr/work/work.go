package jwork

import (
	"jglobal"
	"jmedia"
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
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DeleteUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Register(jpb.CMD_SWIPER_LIST_REQ, swiperList, &jpb.SwiperListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_SWIPER_REQ, uploadSwiper, &jpb.UploadSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_SWIPER_REQ, deleteSwiper, &jpb.DeleteSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_GOOD_LIST_REQ, goodList, &jpb.GoodListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_GOOD_REQ, uploadGood, &jpb.UploadGoodReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_GOOD_REQ, deleteGood, &jpb.DeleteGoodReq{})
	jnet.Rpc.Register(jpb.CMD_IMAGE_REQ, image, &jpb.ImageReq{})
	jnet.Rpc.Register(jpb.CMD_VIDEO_REQ, video, &jpb.VideoReq{})
}

// ------------------------- inside.method -------------------------

// 缓存清理
func deleteUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.DeleteUserReq)
	pack.Data = &jpb.DeleteUserRsp{}
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

// 获取轮播图列表
func swiperList(pack *jglobal.Pack) {
	rsp := &jpb.SwiperListRsp{}
	pack.Cmd = jpb.CMD_SWIPER_LIST_RSP
	pack.Data = rsp
	user := juser.GetUserAnyway(0)
	rsp.MUids = user.Swipers.Data
}

// 上传轮播图
func uploadSwiper(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	req := pack.Data.(*jpb.UploadSwiperReq)
	rsp := &jpb.UploadSwiperRsp{}
	pack.Cmd = jpb.CMD_UPLOAD_SWIPER_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	uids, err := jmedia.Media.Add([]*jpb.Media{req.Media})
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	user0 := juser.GetUserAnyway(0)
	for k, v := range uids {
		user0.AddSwiper(k, v)
		break
	}
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
		})
	})
}

// 删除轮播图
func deleteSwiper(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser.User)
	req := pack.Data.(*jpb.DeleteSwiperReq)
	rsp := &jpb.DeleteSwiperRsp{}
	pack.Cmd = jpb.CMD_DELETE_SWIPER_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	if err := jmedia.Media.Delete([]uint32{req.Uid}); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	user0 := juser.GetUserAnyway(0)
	user0.DelSwiper(req.Uid)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
		})
	})
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
	uids, err := jmedia.Media.Add(req.Good.Medias)
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	req.Good.Medias = nil
	req.Good.MUids = uids
	user0 := juser.GetUserAnyway(0)
	uid, err := user0.GenGoodUid()
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	user0.AddGood(uid, req.Good)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
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
	good := user0.Goods.Data[req.Uid]
	if good == nil {
		rsp.Code = jpb.CODE_PARAM
		return
	}
	uids := []uint32{}
	for k := range good.MUids {
		uids = append(uids, k)
	}
	if err := jmedia.Media.Delete(uids); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	user0.DelGood(req.Uid)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
		})
	})
}

// 下载图片
func image(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.ImageReq)
	rsp := &jpb.ImageRsp{}
	pack.Cmd = jpb.CMD_IMAGE_RSP
	pack.Data = rsp
	image, err := jmedia.Media.GetImage(req.Uid)
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
	} else {
		rsp.Image = image
	}
}

// 下载视频
func video(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.VideoReq)
	rsp := &jpb.VideoRsp{}
	pack.Cmd = jpb.CMD_VIDEO_RSP
	pack.Data = rsp
	video, err := jmedia.Media.GetVideo(req.Uid)
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
	} else {
		rsp.Video = video
	}
}
