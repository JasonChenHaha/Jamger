package juser2

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Basic struct {
	user    *User
	Id      string
	Admin   bool
	Gate    int
	LoginTs int64
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	return &Basic{user: user}
}

func (basic *Basic) load(data bson.M) {
	if v, ok := data["basic"]; ok {
		data = v.(bson.M)
		if v2, ok := data["loginTs"]; ok {
			basic.LoginTs = v2.(int64)
		}
		if _, ok := data["admin"]; ok {
			basic.Admin = true
		}
	}
}

// ------------------------- outside -------------------------

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
