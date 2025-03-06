package juser

import (
	"jdb"
	"jglobal"
	"jlog"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Mongo struct {
	user *User
	*Basic
	*Goods
}

// ------------------------- package -------------------------

func newMongo(user *User) *Mongo {
	return &Mongo{
		user:  user,
		Basic: newBasic(user),
		Goods: newGoods(user),
	}
}

// ------------------------- outside -------------------------

func (mongo *Mongo) Load() *User {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": mongo.user.Uid},
		Project: bson.M{"basic": 1, "goods": 1},
	}
	data := primitive.M{}
	if err := jdb.Mongo.FindOne(in, &data); err != nil {
		jlog.Error(err)
		return nil
	}
	mongo.Basic.load(data)
	mongo.Goods.load(data)
	return mongo.user
}
