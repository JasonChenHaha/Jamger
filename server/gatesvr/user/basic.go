package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Basic struct {
	user  *User
	Id    string
	Pwd   []byte
	SesId uint64
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
		basic.Id = mData["id"].(string)
		basic.Pwd = mData["pwd"].(primitive.Binary).Data
	}
	return basic.user
}

func (basic *Basic) GetSesId() uint64 {
	return basic.SesId
}

func (basic *Basic) SetSesId(id uint64) {
	basic.SesId = id
}
