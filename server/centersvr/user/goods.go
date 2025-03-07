package juser

import (
	"fmt"
	"jpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
				Id:      uint32(v3["id"].(int64)),
				Name:    v3["name"].(string),
				Desc:    v3["desc"].(string),
				Size:    uint32(v3["size"].(int64)),
				Price:   uint32(v3["price"].(int64)),
				ImgType: uint32(v3["imgtype"].(int64)),
				Image:   v3["image"].(primitive.Binary).Data,
			}
			tmp[good.Id] = good
		}
		goods.Data = tmp
	}
}

// ------------------------- outside -------------------------

func (goods *Goods) AddGood(id uint32, good *jpb.Good) {
	good.Id = id
	goods.Data[id] = good
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", id)] = good
	goods.user.UnLock()
}

func (goods *Goods) ModifyGood(good *jpb.Good) {
	goods.Data[good.Id] = good
	goods.user.Lock()
	goods.user.DirtyMongo[fmt.Sprintf("goods.%d", good.Id)] = good
	goods.user.UnLock()
}

func (goods *Goods) DelGood(id uint32) {
	delete(goods.Data, id)
	goods.user.Lock()
	goods.user.DirtyMongo2[fmt.Sprintf("goods.%d", id)] = 1
	goods.user.UnLock()
}
