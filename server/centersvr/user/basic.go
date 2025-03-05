package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Basic struct {
	user    *User
	Id      string
	Gate    int
	LoginTs int64
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	return &Basic{user: user}
}

// ------------------------- outside -------------------------

func (basic *Basic) Load() *User {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": basic.user.Uid},
		Project: bson.M{"basic": 1},
	}
	mData := primitive.M{}
	if err := jdb.Mongo.FindOne(in, &mData); err != nil {
		jlog.Error(err)
		return nil
	}
	if v, ok := mData["basic"]; ok {
		mData = v.(primitive.M)
		if v2, ok2 := mData["loginTs"]; ok2 {
			basic.LoginTs = v2.(int64)
		}
	}
	return basic.user
}

func (basic *Basic) SetGate(gate int) {
	basic.Gate = gate
}

func (basic *Basic) GetGate() int {
	return basic.Gate
}

func (basic *Basic) SetLoginTs() {
	basic.LoginTs = time.Now().Unix()
	basic.user.Lock()
	basic.user.DirtyMongo["basic.loginTs"] = basic.LoginTs
	basic.user.UnLock()
}
