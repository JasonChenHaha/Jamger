package jglobal2

import "jconfig"

var AppId string
var AppSecret string

func Init() {
	AppId = jconfig.GetString("app.id")
	AppId = jconfig.GetString("app.secret")
}
