package juser2

import "go.mongodb.org/mongo-driver/bson"

type Score struct {
	user *User
	Data uint32
}

// ------------------------- package -------------------------

func newScore(user *User) *Score {
	return &Score{user: user}
}

func (sc *Score) load(data bson.M) {
	if v, ok := data["score"]; ok {
		sc.Data = uint32(v.(int64))
	}
}

// ------------------------- outside -------------------------

// 修改积分
func (sc *Score) ModifyScore(score uint32) {
	sc.Data = score
	sc.user.Lock()
	sc.user.DirtyMongo["score"] = score
	sc.user.UnLock()
}
