package juser

import (
	"time"

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

func (basic *Basic) load(data primitive.M) {
	if v, ok := data["basic"]; ok {
		data = v.(primitive.M)
		if v2, ok2 := data["loginTs"]; ok2 {
			basic.LoginTs = v2.(int64)
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
