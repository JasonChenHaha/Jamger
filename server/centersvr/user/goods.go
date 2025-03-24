package juser

import (
	"fmt"
	"jdb"
	"jglobal"
	"jmongo"
	"jpb"

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
		tmp := map[uint32]*jpb.Good{}
		for _, v2 := range v.(bson.M) {
			v3 := v2.(bson.M)
			good := &jpb.Good{
				Uid:    uint32(v3["uid"].(int64)),
				Name:   v3["name"].(string),
				Desc:   v3["desc"].(string),
				Size:   v3["size"].(string),
				Oprice: uint32(v3["oprice"].(int64)),
				Price:  uint32(v3["price"].(int64)),
				Kind:   v3["kind"].(string),
			}
			tmp[good.Uid] = good
		}
		goods.Data = tmp
	}
}

// ------------------------- outside -------------------------

// 生成商品id
func (goods *Goods) GenGoodUid() (uint32, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": int64(0)},
		Update:  bson.M{"$inc": bson.M{"guidc": int64(1)}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"guidc": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return 0, err
	}
	return uint32(out["guidc"].(int64)), nil
}

// 添加商品
func (goods *Goods) AddGood(uid uint32, good *jpb.Good) {
	good.Uid = uid
	goods.Data[uid] = good
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", uid)] = good
	goods.user.UnLock()
}

// 修改商品
func (goods *Goods) ModifyGood(good *jpb.Good) {
	goods.Data[good.Uid] = good
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", good.Uid)] = good
	goods.user.UnLock()
}

// 下架商品
func (goods *Goods) DelGood(uid uint32) {
	delete(goods.Data, uid)
	goods.user.Lock()
	goods.user.DirtyMongo2[fmt.Sprintf("goods.%d", uid)] = 1
	goods.user.UnLock()
}
