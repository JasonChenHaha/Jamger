package juser

import (
	"time"
)

type Basic struct {
	dm      map[string]any
	dr      map[string]any
	Id      string
	LoginTs int64
}

// ------------------------- package -------------------------

func newBasic(user *User, mData map[string]any, rData map[string]string) *Basic {
	basic := &Basic{
		dm: user.Base.DirtyMongo,
		dr: user.Base.DirtyRedis,
	}
	if sub, ok := mData["basic"]; ok {
		mData = sub.(map[string]any)
		basic.LoginTs = mData["loginTs"].(int64)
	}
	return basic
}

// ------------------------- outside -------------------------

func (basic *Basic) SetLoginTs() {
	basic.LoginTs = time.Now().Unix()
	basic.dm["basic.loginTs"] = basic.LoginTs
}
