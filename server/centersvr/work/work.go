package jwork

import (
	"jaddress"
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
	jnet.Rpc.Register(jpb.CMD_DEL_USER, deleteUser, &jpb.DeleteUserReq{})
	jnet.Rpc.Register(jpb.CMD_LOGIN_REQ, login, &jpb.LoginReq{})
	jnet.Rpc.Register(jpb.CMD_STATUS_REQ, status, &jpb.StatusReq{})
	jnet.Rpc.Register(jpb.CMD_RECORD_REQ, record, &jpb.RecordReq{})
	jnet.Rpc.Register(jpb.CMD_ADD_RECORD_REQ, addRecord, &jpb.AddRecordReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_RECORD_REQ, modifyRecord, &jpb.ModifyRecordReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_RECORD_REQ, deleteRecord, &jpb.DeleteRecordReq{})
	jnet.Rpc.Register(jpb.CMD_SWIPER_LIST_REQ, swiperList, &jpb.SwiperListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_SWIPER_REQ, uploadSwiper, &jpb.UploadSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_SWIPER_REQ, deleteSwiper, &jpb.DeleteSwiperReq{})
	jnet.Rpc.Register(jpb.CMD_GOOD_LIST_REQ, goodList, &jpb.GoodListReq{})
	jnet.Rpc.Register(jpb.CMD_UPLOAD_GOOD_REQ, uploadGood, &jpb.UploadGoodReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_GOOD_REQ, modifyGood, &jpb.ModifyGoodReq{})
	jnet.Rpc.Register(jpb.CMD_DELETE_GOOD_REQ, deleteGood, &jpb.DeleteGoodReq{})
	jnet.Rpc.Register(jpb.CMD_MODIFY_KIND_REQ, modifyKind, &jpb.ModifyKindReq{})
	jnet.Rpc.Register(jpb.CMD_IMAGE_REQ, image, &jpb.ImageReq{})
	jnet.Rpc.Register(jpb.CMD_VIDEO_REQ, video, &jpb.VideoReq{})
	jnet.Rpc.Register(jpb.CMD_ADDRESS_REQ, address, &jpb.AddressReq{})
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

// 获取状态
func status(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	rsp := &jpb.StatusRsp{}
	pack.Cmd = jpb.CMD_STATUS_REQ
	pack.Data = rsp
	count, all, score, done := uint32(0), uint32(0), uint32(0), false
	records := user.Record.Data
	for i := len(records) - 1; i >= 0; i-- {
		if records[i].Score > 0 {
			if !done {
				count++
				score += records[i].Score
			}
			all++
		} else {
			done = true
		}
	}
	rsp.Score = score
	rsp.All = all
	rsp.Count = count
	if count > 0 {
		rsp.Ave = score / count
	}
	rsp.Progress = jglobal.Min(count*100/uint32(jconfig.GetInt("good.free")), 100)
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
	user0 := juser2.GetUserAnyway(0)
	if err := user0.AddSwiper(req.Media); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
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
	user0 := juser2.GetUserAnyway(0)
	if _, ok := user0.Swipers.Data[req.Uid]; !ok {
		rsp.Code = jpb.CODE_PARAM
		return
	}
	if err := user0.DelSwiper(req.Uid); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
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
	user0 := juser2.GetUserAnyway(0)
	if req.Good.Uid != 0 {
		if err := user0.DelGood(req.Good.Uid); err != nil {
			rsp.Code = jpb.CODE_SVR_ERR
			return
		}
	}
	if err := user0.AddGood(req.Good); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
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
	if user0.Goods.Data[req.Good.Uid] == nil {
		rsp.Code = jpb.CODE_PARAM
		return
	}
	if err := user0.ModifyGood(req.Good); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
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
	if err := user0.DelGood(req.Uid); err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
		return
	}
	jschedule.DoAt(5*time.Second, func(args ...any) {
		jnet.BroadcastToGroup(jglobal.GRP_CENTER, &jglobal.Pack{
			Cmd:  jpb.CMD_DEL_USER,
			Data: &jpb.DeleteUserReq{Uid: 0},
		})
	})
}

// 修改品类
func modifyKind(pack *jglobal.Pack) {
	user := pack.Ctx.(*juser2.User)
	req := pack.Data.(*jpb.ModifyKindReq)
	rsp := &jpb.ModifyKindRsp{}
	pack.Cmd = jpb.CMD_MODIFY_KIND_RSP
	pack.Data = rsp
	if !user.Admin {
		rsp.Code = jpb.CODE_DENY
		return
	}
	good := &jpb.Good{Kind: req.New}
	user0 := juser2.GetUserAnyway(0)
	for _, v := range user0.Goods.Data {
		if v.Kind == req.Old {
			good.Uid = v.Uid
			user0.ModifyGood(good)
		}
	}
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
	jlog.Debugf("download image %d", req.Uid)
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
	jlog.Debugf("download video %d", req.Uid)
	rsp := &jpb.VideoRsp{}
	pack.Cmd = jpb.CMD_VIDEO_RSP
	pack.Data = rsp
	video, err := jmedia.Media.GetVideo(req.Uid)
	if err != nil {
		rsp.Code = jpb.CODE_SVR_ERR
	} else {
		size := uint32(len(video))
		rsp.Size = size
		if req.End == 0 {
			return
		} else if req.End == 1 {
			req.End = req.Start + uint32(jconfig.GetInt("media.video.len"))
			rsp.Video = video[req.Start:jglobal.Min(req.End+1, size)]
		} else if req.Start > req.End || req.Start >= size {
			rsp.Code = jpb.CODE_PARAM
			return
		} else {
			rsp.Video = video[req.Start:jglobal.Min(req.End+1, size)]
		}
	}
}

// 获取门店地址
func address(pack *jglobal.Pack) {
	rsp := &jpb.AddressRsp{}
	pack.Cmd = jpb.CMD_ADDRESS_RSP
	pack.Data = rsp
	rsp.Addrs = jaddress.Addr.Get()
}
