package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
)

type Basic struct {
	user *User
	Id   string
	Pwd  []byte
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	basic := &Basic{user: user}
	return basic
}

func (basic *Basic) Load() {
	if basic.Pwd != nil {
		return
	}
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": basic.user.Uid},
		Project: bson.M{"auth": 1},
	}
	mData := map[string]any{}
	if err := jdb.Mongo.FindOne(in, &mData); err != nil {
		jlog.Error(err)
		return
	}
	if v, ok := mData["auth"]; ok {
		mData = v.(map[string]any)
		basic.Id = mData["id"].(string)
		basic.Pwd = mData["pwd"].([]byte)
	}
}

// ------------------------- outside -------------------------

func (basic *Basic) SetId(id string) {
	basic.Id = id
	basic.user.DirtyMongo["basic.id"] = id
}

func (basic *Basic) SetPwd(pwd []byte) {
	basic.Pwd = pwd
	basic.user.DirtyMongo["basic.pwd"] = pwd
}
