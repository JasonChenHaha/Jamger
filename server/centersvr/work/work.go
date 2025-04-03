package jwork

import (
	"jconfig"
	"jglobal"
	"jlog"
	"jmedia"
	"jnet"
	"jpb"
	"jrpc"
	"jschedule"
	"juser2"
	"time"
)

// ------------------------- outside -------------------------

func Init() {
	jrpc.Rpc.Connect(jglobal.GRP_CENTER)
	jrpc.Rpc.Connect(jglobal.GRP_GATE)
	jnet.Rpc.SetCodec(rpcEncode, rpcDecode)
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DeleteUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Register(jpb.CMD_RECORD_REQ, record, &jpb.RecordReq{})
	jnet.Rpc.Register(jpb.CMD_ADD_RECORD_REQ, addRecord, &jpb.AddRecordReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_RECORD_REQ, modifyRecord, &jpb.ModifyRecordReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_RECORD_REQ, deleteRecord, &jpb.DeleteRecoredReq{})
	jnet.Rpc.Register(jpb.CMD_SWIPER_LIST_REQ, swiperList, &jpb.SwiperListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_SWIPER_REQ, uploadSwiper, &jpb.UploadSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_SWIPER_REQ, deleteSwiper, &jpb.DeleteSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_GOOD_LIST_REQ, goodList, &jpb.GoodListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_GOOD_REQ, uploadGood, &jpb.UploadGoodReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_GOOD_REQ, modifyGood, &jpb.ModifyGoodReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_GOOD_REQ, deleteGood, &jpb.DeleteGoodReq{})
	jnet.Rpc.Register(jpb.CMD_IMAGE_REQ, image, &jpb.ImageReq{})
	jnet.Rpc.Register(jpb.CMD_VIDEO_REQ, video, &jpb.VideoReq{})
}

// ------------------------- inside.method -------------------------

// 缓存清理
func deleteUser(pack *jglobal.Pack) {
	req := pack.Data.(*jpb.DeleteUserReq)
	pack.Data = &jpb.DeleteUserRsp{}
	if user := juser2.GetUser(req.Uid); user != nil {
		user.Destory()
	}
}

// 登录
func login(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	rsp := &jpb.LoginRsp{}
	pack.Cmd = jpb.CMD_LOGIN_RSP
	pack.Data = rsp
	user.SetLoginTs()
}

// 获取记录
func record(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.RecordReq)
	rsp := &jpb.RecordRsp{}
	pack.Cmd = jpb.CMD_RECORD_RSP
	pack.Data = rsp
	if req.Uid != 0 {
		if !user.Admin {
			rsp.Code = jpb.CODE_DENY
			return
		}
		user2 := juser2.GetUserAnyway(req.Uid)
		if !user2.Exist {
			rsp.Code = jpb.CODE_USER_NIL
			return
		}
		rsp.Records = user2.Record.Data
	} else {
		rsp.Records = user.Record.Data
	}
}

// 增加记录
func addRecord(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.AddRecordReq)
	rsp := &jpb.AddRecordRsp{}
	pack.Cmd = jpb.CMD_ADD_RECORD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user2 := juser2.GetUserAnyway(req.Uid)
	if !user2.Exist {
		rsp.Code = jpb.CODE_USER_NIL
		return
	}
	user2.AddRecord(req.Record)
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: req.Uid},
		})
	})
}

// 修改记录
func modifyRecord(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.ModifyRecordReq)
	rsp := &jpb.ModifyRecordRsp{}
	pack.Cmd = jpb.CMD_MODIFY_RECORD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user2 := juser2.GetUserAnyway(req.Uid)
	if !user2.Exist {
		rsp.Code = jpb.CODE_USER_NIL
		return
	}
	if !user2.ModifyRecord(req.Index, req.Record) {
		rsp.Code = jpb.CODE_PARAM
		return
	}
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: req.Uid},
		})
	})
}

// 删除记录
func deleteRecord(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.DeleteRecordReq)
	rsp := &jpb.DeleteRecordRsp{}
	pack.Cmd = jpb.CMD_DELETE_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user2 := juser2.GetUserAnyway(req.Uid)
	if !user2.DeleteRecord(req.Index) {
		rsp.Code = jpb.CODE_PARAM
		return
	}
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: req.Uid},
		})
	})
}

// 获取轮播图列表
func swiperList(pack *jglobal.Pack) {
	rsp := &jpb.SwiperListRsp{}
	pack.Cmd = jpb.CMD_SWIPER_LIST_RSP
	pack.Data = rsp
	user := juser2.GetUserAnyway(0)
	rsp.MUids = user.Swipers.Data
}

// 上传轮播图
func uploadSwiper(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
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
	user0 := juser2.GetUserAnyway(0)
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
	user := pack.Ctx.(*juser2.User)
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
	user0 := juser2.GetUserAnyway(0)
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
	user := juser2.GetUserAnyway(0)
	for _, v := range user.Goods.Data {
		rsp.Goods = append(rsp.Goods, v)
	}
}

// 上传商品
func uploadGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
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
	user0 := juser2.GetUserAnyway(0)
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

// 修改商品
func modifyGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.ModifyGoodReq)
	rsp := &jpb.ModifyGoodRsp{}
	pack.Cmd = jpb.CMD_MODIFY_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user0 := juser2.GetUserAnyway(0)
	if req.Good.Size != "" {
		good := user0.Goods.Data[req.Good.Uid]
		good.Size = req.Good.Size
		user0.ModifyGood(req.Good.Uid, good)
	} else {
		user0.DelGood(req.Good.Uid)
	}
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
		})
	})
}

// 下架商品
func deleteGood(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.DeleteGoodReq)
	rsp := &jpb.DeleteGoodRsp{}
	pack.Cmd = jpb.CMD_DELETE_GOOD_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	user0 := juser2.GetUserAnyway(0)
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
	jlog.Debugf("image size: %d", len(rsp.Image))
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
		if req.End == 0 {
			req.End = req.Start + uint32(jconfig.GetInt("video.len"))
		}
		size := uint32(len(video))
		jlog.Debugf("video size: %d", size)
		if req.Start > req.End || req.Start >= size {
			rsp.Code = jpb.CODE_PARAM
			return
		}
		rsp.Size = size
		rsp.Video = video[req.Start:jglobal.Min(req.End+1, size)]
	}
}
