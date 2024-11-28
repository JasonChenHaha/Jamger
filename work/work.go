package jwork

import (
	jdb "jamger/db"
	jdebug "jamger/debug"
	jlog "jamger/log"
	jnet "jamger/net"
	jkcp "jamger/net/kcp"
	jtcp "jamger/net/tcp"
	jweb "jamger/net/web"

	"go.mongodb.org/mongo-driver/bson"
)

type DDD struct {
	Uin  uint32
	Name string
}

func Run() {
	jnet.Tcp.RegisterHandler(1, cb1)
	jnet.Kcp.RegisterHandler(2, cb2)
	jnet.Web.RegisterHandler(1, cb3)

	o := &DDD{}
	jdb.Mongo.FindOne("test", bson.D{{Key: "uin", Value: 0}}, o)
	jlog.Debug(o)
	oo := []*DDD{}
	jdb.Mongo.FindMany("test", bson.D{{Key: "uin", Value: 0}}, &oo)
	jlog.Debug(oo)
	oo = []*DDD{}
	jdb.Mongo.FindMany("test", bson.D{{Key: "uin", Value: bson.D{{Key: "$gte", Value: 0}}}}, &oo)
	jlog.Debug(oo[1])
	o = &DDD{1, "1"}
	jdb.Mongo.InsertOne("test", o)
}

func cb1(id uint64, pack *jtcp.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jnet.Tcp.Send(id, 1, []byte("ok!"))
}

func cb2(id uint64, pack *jkcp.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jnet.Kcp.Send(id, 1, []byte("ok!"))
}

func cb3(id uint64, pack *jweb.Pack) {
	jlog.Debug(jdebug.StructToString(pack))
	jnet.Kcp.Send(id, 1, []byte("ok!"))
}
