package juser2

import (
	"jdb"
	"jglobal"
	"jmongo"

	"go.mongodb.org/mongo-driver/bson"
)

type Mongo struct {
	user *User
	*Basic
	*Score
	*Swipers
	*Goods
	Exist bool
}

// ------------------------- package -------------------------

func newMongo(user *User) *Mongo {
	return &Mongo{
		user:    user,
		Basic:   newBasic(user),
		Score:   newScore(user),
		Swipers: newSwipers(user),
		Goods:   newGoods(user),
	}
}

// ------------------------- outside -------------------------

func (mongo *Mongo) Load() *User {
	in := &jmongo.Input{
		Col:     jglobal.MONGO_USER,
		Filter:  bson.M{"_id": mongo.user.Uid},
		Project: bson.M{"basic": 1, "swipers": 1, "goods": 1, "score": 1},
	}
	data := bson.M{}
	if err := jdb.Mongo.FindOne(in, &data); err != nil {
		return mongo.user
	}
	mongo.Exist = true
	mongo.Basic.load(data)
	mongo.Score.load(data)
	mongo.Swipers.load(data)
	mongo.Goods.load(data)
	return mongo.user
}
