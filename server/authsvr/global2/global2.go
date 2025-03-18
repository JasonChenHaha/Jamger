package jglobal2

import "jconfig"

var AppId string
var AppSecret string

func Init() {
	AppId = jconfig.GetString("app.id")
	AppSecret = jconfig.GetString("app.secret")
}
