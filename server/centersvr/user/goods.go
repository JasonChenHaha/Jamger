package juser

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Good struct {
	Id      uint32
	Name    string
	Desc    string
	Size    uint32
	Price   uint32
	ImgType uint32
	Image   []byte
}

type Goods struct {
	user *User
	Data []*Good
}

// ------------------------- package -------------------------

func newGoods(user *User) *Goods {
	return &Goods{
		user: user,
		Data: []*Good{},
	}
}

func (goods *Goods) load(data primitive.M) {
	if v, ok := data["goods"]; ok {
		tmp := []*Good{}
		for _, v2 := range v.(primitive.A) {
			v3 := v2.(primitive.M)
			good := &Good{
				Id:      uint32(v3["id"].(int32)),
				Name:    v3["name"].(string),
				Desc:    v3["desc"].(string),
				Size:    uint32(v3["size"].(int32)),
				Price:   uint32(v3["price"].(int32)),
				ImgType: uint32(v3["imgType"].(int32)),
				Image:   v3["image"].(primitive.Binary).Data,
			}
			tmp = append(tmp, good)
		}
		goods.Data = tmp
	}
}

// ------------------------- outside -------------------------
