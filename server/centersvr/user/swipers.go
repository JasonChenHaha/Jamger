package juser

import (
	"fmt"
	"jglobal"

	"go.mongodb.org/mongo-driver/bson"
)

type Swipers struct {
	user *User
	Data map[uint32]uint32
}

// ------------------------- package -------------------------

func newSwipers(user *User) *Swipers {
	return &Swipers{
		user: user,
		Data: map[uint32]uint32{},
	}
}

func (sp *Swipers) load(data bson.M) {
	if v, ok := data["swipers"]; ok {
		tmp := map[uint32]uint32{}
		for k, ty := range v.(bson.M) {
			tmp[jglobal.Atoi[uint32](k)] = uint32(ty.(int64))
		}
		sp.Data = tmp
	}
}

// ------------------------- outside -------------------------

// 添加轮播图
func (sp *Swipers) AddSwiper(uid uint32, ty uint32) {
	sp.Data[uid] = ty
	sp.user.Lock()
	sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", uid)] = ty
	sp.user.UnLock()
}

// 删除轮播图
func (sp *Swipers) DelSwiper(uid uint32) {
	delete(sp.Data, uid)
	sp.user.Lock()
	sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", uid)] = nil
	sp.user.UnLock()
}
