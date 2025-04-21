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
	goo := goods.Data[good.Uid]
	if len(good.Name) > 0 {
		if good.Name == "." {
			goo.Name = ""
		} else {
			goo.Name = good.Name
		}
	}
	if len(good.Desc) > 0 {
		if good.Desc == "." {
			goo.Desc = ""
		} else {
			goo.Desc = good.Desc
		}
	}
	if len(good.Size) > 0 {
		if good.Size == "." {
			goo.Expire = time.Now().Unix() + int64(jconfig.GetInt("good.expire"))
			goo.Size = ""
		} else {
			goo.Expire = 0
			goo.Size = good.Size
		}
	}
	if good.Oprice != 0 {
		if good.Oprice == 9999 {
			goo.Oprice = 0
		} else {
			goo.Oprice = good.Oprice
		}
	}
	if good.Price != 0 {
		if good.Price == 9999 {
			goo.Price = 0
		} else {
			goo.Price = good.Price
		}
	}
	if len(good.Kind) > 0 {
		if good.Kind == "." {
			goo.Kind = ""
		} else {
			goo.Kind = good.Kind
		}
	}
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", good.Uid)] = goo
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
