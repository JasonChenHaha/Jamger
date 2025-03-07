package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
)

type Mongo struct {
	user *User
	*Basic
}

// ------------------------- package -------------------------

func newMongo(user *User) *Mongo {
	return &Mongo{
		user:  user,
		Basic: newBasic(user),
	}
}

// ------------------------- outside -------------------------

func (mongo *Mongo) Load() *User {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": mongo.user.Uid},
		Project: bson.M{"basic": 1},
	}
	data := bson.M{}
	if err := jdb.Mongo.FindOne(in, &data); err != nil {
		jlog.Error(err)
		return nil
	}
	mongo.Basic.load(data)
	return mongo.user
}
