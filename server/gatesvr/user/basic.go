package juser

import (
	"bytes"
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Basic struct {
	user *User
	Id   string
	Pwd  []byte
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
		basic.Id = mData["id"].(string)
		basic.Pwd = mData["pwd"].(primitive.Binary).Data
	}
}

// ------------------------- outside -------------------------

func (basic *Basic) SetId(id string) {
	if basic.Id != id {
		basic.Id = id
		basic.user.DirtyMongo["basic.id"] = id
	}
}

func (basic *Basic) SetPwd(pwd []byte) {
	if !bytes.Equal(basic.Pwd, pwd) {
		basic.Pwd = pwd
		basic.user.DirtyMongo["basic.pwd"] = pwd
	}
}
