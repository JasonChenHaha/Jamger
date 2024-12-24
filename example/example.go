package jexample

import (
	"context"
	"fmt"
	"jdb"
	"jdebug"
	"jevent"
	"jglobal"
	"jkcp"
	"jlog"
	"jmongo"
	"jnet"
	pb "jpb"
	"jrpc"
	"jschedule"
	"jtcp"
	"jweb"
	"reflect"
	"time"

	"github.com/nsqio/go-nsq"
	"go.mongodb.org/mongo-driver/bson"
)

type DDD struct {
	Uin  uint32
	Name string
}

type GateServer struct {
	pb.GateServer
}

func (svr *GateServer) SayHello(ctx context.Context, req *pb.RequestGate) (*pb.GateResponse, error) {
	return &pb.GateResponse{
		Message: fmt.Sprintf("hello %s, this is %s", req.GetName(), jglobal.SERVER),
	}, nil
}

type GameServer struct {
	pb.GameServer
}

func (svr *GameServer) SayHello(ctx context.Context, req *pb.RequestGame) (*pb.GameResponse, error) {
	return &pb.GameResponse{
		Message: fmt.Sprintf("hello %s, this is %s", req.GetName(), jglobal.SERVER),
	}, nil
}

// ------------------------- inside -------------------------

func Init() {
	// network()
	// mongo()
	// redis()
	// schedule()
	// event()
	// rpc()
}

func network() {
	jnet.Tcp.Register(1, func(id uint64, pack *jtcp.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Tcp.Send(id, 1, []byte("ok!"))
	})
	jnet.Kcp.Register(2, func(id uint64, pack *jkcp.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Kcp.Send(id, 1, []byte("ok!"))
	})
	jnet.Web.Register(1, func(id uint64, pack *jweb.Pack) {
		jlog.Debug(jdebug.StructToString(pack))
		jnet.Kcp.Send(id, 1, []byte("ok!"))
	})
}

func mongo() {
	in := &jmongo.Input{
		Col:     "test",
		Filter:  bson.M{"uin": 1},
		Project: bson.M{"name": 1},
	}
	var out any
	in = &jmongo.Input{
		Col: "test",
	}
	count, _ := jdb.Mongo.EstimatedDocumentCount(in)
	jlog.Debug(count)
	count, _ = jdb.Mongo.CountDocuments(in)
	jlog.Debug(count)
	out = &DDD{}
	jdb.Mongo.FindOne(in, out)
	jlog.Debug(out)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 0},
		Sort:   bson.M{"uin": 1},
		Limit:  1,
	}
	out = &[]*DDD{}
	jdb.Mongo.FindMany(in, out)
	jlog.Debug(out)
	in = &jmongo.Input{
		Col:    "test",
		Insert: &DDD{2, "2"},
	}
	jdb.Mongo.InsertOne(in)
	in = &jmongo.Input{
		Col: "test",
		InsertMany: []any{
			&DDD{Uin: 3, Name: "3"},
			&DDD{Uin: 4, Name: "4"},
		},
	}
	jdb.Mongo.InsertMany(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 2},
		Update: bson.M{"$set": bson.M{"name": "2"}},
	}
	jdb.Mongo.UpdateOne(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 2},
		Update: bson.M{"$set": bson.M{"name": "2"}},
	}
	jdb.Mongo.UpdateMany(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 4},
	}
	jdb.Mongo.DeleteOne(in)
	in = &jmongo.Input{
		Col:    "test",
		Filter: bson.M{"uin": 3},
	}
	jdb.Mongo.DeleteMany(in)
}

func redis() {
	res, _ := jdb.Redis.Do("SET", "jamger", "123", "EX", 3)
	jlog.Debug(res)
	res, _ = jdb.Redis.Do("GET", "jamger")
	jlog.Debug(reflect.TypeOf(res).Kind())
	scr := `
		local value = redis.call('GET', KEYS[1])
		return value
	`
	res, _ = jdb.Redis.DoScript(scr, []string{"jamger"})
	jlog.Debug(res)
}

func schedule() {
	id := jschedule.DoEvery(1*time.Second, func() {
		jlog.Debug("doevery")
	})
	jschedule.Stop(id)
	id = jschedule.DoCron("* * * * * *", func() {
		jlog.Debug("docron")
		jevent.LocalPublish(jevent.EVENT_TEST_1, nil)
		if jglobal.SERVER == "jamger1" {
			jevent.RemotePublish(jevent.EVENT_TEST_2, []byte("recv remote event"))
		}
	})
	jschedule.Stop(id)
	id = jschedule.DoAt(20*time.Second, func() {
		jlog.Debug("doat")

	})
	jschedule.Stop(id)
}

func event() {
	jevent.LocalRegister(jevent.EVENT_TEST_1, func(context any) {
		jlog.Debug("recv local event")
	})
	jevent.LocalRegister(jevent.EVENT_TEST_1, func(context any) {
		jlog.Debug("recv local event")
	})
	jevent.RemoteRegister(jevent.EVENT_TEST_2, func(msg *nsq.Message) error {
		jlog.Debug(jglobal.SERVER, string(msg.Body))
		return nil
	})
}

func rpc() {
	f := func(target any) {
		res, err := target.(pb.GateClient).SayHello(context.Background(), &pb.RequestGate{
			Name: jglobal.SERVER,
		})
		if err != nil {
			jlog.Error(err)
		} else {
			jlog.Debug(res.Message)
		}
	}

	if jglobal.GROUP == jglobal.SVR_ALPHA {
		jrpc.Server(&pb.Game_ServiceDesc, &GameServer{})
		jrpc.Connect(jglobal.SVR_BETA, pb.NewGateClient)
		i := 0
		jschedule.DoCron("*/5 * * * * *", func() {
			if target := jrpc.GetTarget(jglobal.SVR_BETA, "gate-01"); target != nil {
				f(target)
			}
			if target := jrpc.GetFixHashTarget(jglobal.SVR_BETA, i); target != nil {
				f(target)
			}
			if target := jrpc.GetRoundRobinTarget(jglobal.SVR_BETA); target != nil {
				f(target)
			}
			if target := jrpc.GetConsistentHashTarget(jglobal.SVR_BETA, i); target != nil {
				f(target)
			}
			i++
		})
	}
	if jglobal.GROUP == jglobal.SVR_BETA {
		jrpc.Server(&pb.Gate_ServiceDesc, &GateServer{})
		jrpc.Connect(jglobal.SVR_ALPHA, pb.NewGameClient)
	}
}
