package juser2

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Basic struct {
	user    *User
	Id      string
	Pwd     []byte
	Admin   bool
	SesType int
	SesId   uint64
}

// ------------------------- package -------------------------

func newBasic(user *User) *Basic {
	return &Basic{user: user}
}

func (basic *Basic) load(data bson.M) {
	if v, ok := data["basic"]; ok {
		data = v.(bson.M)
		basic.Id = data["id"].(string)
		basic.Pwd = data["pwd"].(primitive.Binary).Data
		if _, ok := data["admin"]; ok {
			basic.Admin = true
		}
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
