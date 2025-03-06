package juser

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Basic struct {
	user    *User
	Id      string
	Pwd     []byte
	SesType int
	SesId   uint64
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	return &Basic{user: user}
}

func (basic *Basic) load(data primitive.M) {
	if v, ok := data["basic"]; ok {
		data = v.(primitive.M)
		basic.Id = data["id"].(string)
		basic.Pwd = data["pwd"].(primitive.Binary).Data
	}
}

// ------------------------- outside -------------------------

func (basic *Basic) GetSesId() (int, uint64) {
	return basic.SesType, basic.SesId
}

func (basic *Basic) SetSesId(tp int, id uint64) {
	basic.SesType = tp
	basic.SesId = id
}
