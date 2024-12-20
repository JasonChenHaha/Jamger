package jwork

import (
	"context"
	"jglobal"
	pb "jpb"
	"jrpc"
	"jschedule"
)

type Work struct {
	pb.GameServer
}

// ------------------------- outside -------------------------

func Init() {
	jrpc.Server(&pb.Game_ServiceDesc, &Work{})
	jrpc.Connect(jglobal.SVR_GATE, pb.NewGateClient)
	jschedule.DoEvery("*/5 * * * * *", func() {
		_, _ = jrpc.GetConsistentHashTarget("gate", 0).(pb.GateClient)
	})
}

func (svr *Work) SayHello(ctx context.Context, req *pb.RequestGame) (*pb.GameResponse, error) {
	return &pb.GameResponse{
		Message: "hello, " + req.GetName(),
	}, nil
}
