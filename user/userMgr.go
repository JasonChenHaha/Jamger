package juser

import "sync"

var userMgr *UserMgr

type UserMgr struct {
	user sync.Map
}

// ------------------------- outside -------------------------

func Init() {
	userMgr = &UserMgr{}
}

// ------------------------- inside -------------------------

func GetUser(uid uint32) *User {
	if v, ok := userMgr.user.Load(uid); ok {
		return v.(*User)
	}
	return nil
}
