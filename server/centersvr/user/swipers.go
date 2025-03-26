package juser

import (
	"fmt"
	"jdb"
	"jglobal"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Swipers struct {
	user *User
	Data map[uint32]struct{}
}

// ------------------------- package -------------------------

func newSwipers(user *User) *Swipers {
	return &Swipers{
		user: user,
		Data: map[uint32]struct{}{},
	}
}

func (sp *Swipers) load(data bson.M) {
	if v, ok := data["swipers"]; ok {
		tmp := map[uint32]struct{}{}
		for k := range v.(bson.M) {
			tmp[jglobal.Atoi[uint32](k)] = struct{}{}
		}
		sp.Data = tmp
	}
}

// ------------------------- outside -------------------------

// 生成轮播图id
func (goods *Goods) GenSwiperUid() (uint32, error) {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": int64(0)},
		Update:  bson.M{"$inc": bson.M{"iuidc": int64(1)}},
		Upsert:  true,
		RetDoc:  options.After,
		Project: bson.M{"iuidc": 1},
	}
	out := bson.M{}
	if err := jdb.Mongo.FindOneAndUpdate(in, &out); err != nil {
		return 0, err
	}
	return uint32(out["iuidc"].(int64)), nil
}

// 添加轮播图
func (sp *Swipers) AddSwiper(uid uint32) {
	sp.Data[uid] = struct{}{}
	sp.user.Lock()
	sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", uid)] = 0
	sp.user.UnLock()
}

// 删除轮播图
func (sp *Swipers) DelSwiper(uid uint32) {
	delete(sp.Data, uid)
	sp.user.Lock()
	sp.user.DirtyMongo[fmt.Sprintf("swipers.%d", uid)] = nil
	sp.user.UnLock()
}
