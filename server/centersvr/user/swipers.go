package juser2

import (
	"fmt"
	"jglobal"
	"jmedia"
	"jpb"

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
func (sp *Swipers) AddSwiper(media *jpb.Media) error {
	uids, err := jmedia.Media.Add([]*jpb.Media{media})
	if err != nil {
		return err
	}
	sp.user.Lock()
	for k, v := range uids {
		sp.Data[k] = v
		sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", k)] = v
	}
	sp.user.UnLock()
	return nil
}

// 删除轮播图
func (sp *Swipers) DelSwiper(uid uint32) error {
	if err := jmedia.Media.Delete([]uint32{uid}); err != nil {
		return err
	}
	delete(sp.Data, uid)
	sp.user.Lock()
	sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", uid)] = nil
	sp.user.UnLock()
	return nil
}
