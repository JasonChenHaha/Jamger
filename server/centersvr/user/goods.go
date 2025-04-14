package juser2

import (
	"fmt"
	"jconfig"
	"jdb"
	"jglobal"
	"jmedia"
	"jmongo"
	"jpb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Goods struct {
	user *User
	Data map[uint32]*jpb.Good
}

// ------------------------- package -------------------------

func newGoods(user *User) *Goods {
	return &Goods{
		user: user,
		Data: map[uint32]*jpb.Good{},
	}
}

func (goods *Goods) load(data bson.M) {
	if v, ok := data["goods"]; ok {
		now := time.Now().Unix()
		tmp, tmp2 := map[uint32]*jpb.Good{}, []uint32{}
		for _, v2 := range v.(bson.M) {
			v3 := v2.(bson.M)
			good := &jpb.Good{
				Uid:    uint32(v3["uid"].(int64)),
				Name:   v3["name"].(string),
				Desc:   v3["desc"].(string),
				Size:   v3["size"].(string),
				Oprice: uint32(v3["oprice"].(int64)),
				Price:  uint32(v3["price"].(int64)),
				MUids:  map[uint32]uint32{},
				Kind:   v3["kind"].(string),
			}
			mUids := v3["muids"].(bson.M)
			for uid, ty := range mUids {
				good.MUids[jglobal.Atoi[uint32](uid)] = uint32(ty.(int64))
			}
			if expire := v3["expire"].(int64); 0 < expire && expire <= now {
				tmp2 = append(tmp2, good.Uid)
			}
			tmp[good.Uid] = good
		}
		goods.Data = tmp
		for _, uid := range tmp2 {
			// 过期清理
			goods.DelGood(uid)
		}
	}
}

// ------------------------- outside -------------------------

// 添加商品
func (goods *Goods) AddGood(good *jpb.Good) error {
	uids, err := jmedia.Media.Add(good.Medias)
	if err != nil {
		return err
	}
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": int64(0)},
		Update:  bson.M{"$inc": bson.M{"guidc": int64(1)}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"guidc": 1},
	}
	out := bson.M{}
	if err = jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return err
	}
	uid := uint32(out["guidc"].(int64))
	good.Uid = uid
	good.MUids = uids
	good.Medias = nil
	goods.Data[uid] = good
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", uid)] = good
	goods.user.UnLock()
	return nil
}

// 修改商品
func (goods *Goods) ModifyGood(good *jpb.Good) {
	good2 := goods.Data[good.Uid]
	good2.Size = good.Size
	if good.Size == "" {
		good2.Expire = time.Now().Unix() + int64(jconfig.GetInt("good.expire"))
	} else {
		good2.Expire = 0
	}
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", good.Uid)] = good2
	goods.user.UnLock()
}

// 下架商品
func (goods *Goods) DelGood(uid uint32) error {
	good := goods.Data[uid]
	uids := []uint32{}
	for k := range good.MUids {
		uids = append(uids, k)
	}
	if err := jmedia.Media.Delete(uids); err != nil {
		return err
	}
	delete(goods.Data, uid)
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", uid)] = nil
	goods.user.UnLock()
	return nil
}
