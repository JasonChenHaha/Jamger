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
	LoginTs int64
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	basic := &Basic{user: user}
	basic.load()
	return basic
}

func (basic *Basic) load() {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": basic.user.Uid},
		Project: bson.M{"basic": 1},
	}
	mData := primitive.M{}
	if err := jdb.Mongo.FindOne(in, &mData); err != nil {
		jlog.Error(err)
		return
	}
	if v, ok := mData["basic"]; ok {
		mData = v.(primitive.M)
		if v2, ok2 := mData["loginTs"]; ok2 {
			basic.LoginTs = v2.(int64)
		}
	}
}

// ------------------------- outside -------------------------

func (basic *Basic) SetLoginTs() {
	basic.LoginTs = time.Now().Unix()
	basic.user.DirtyMongo["basic.loginTs"] = basic.LoginTs
}
