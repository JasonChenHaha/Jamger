package jwork

import (
	"context"
	"jglobal"
	pb "jpb"
	"jrpc"
)

type Work struct {
	pb.GameServer
}

// ------------------------- outside -------------------------

func Init() {
	jrpc.Server(&pb.Game_ServiceDesc, &Work{})
	jrpc.Connect(jglobal.SVR_GATE, pb.NewGateClient)
	// time.Sleep(5 * time.Second)
	// target := jrpc.GetTarget("gate", "gate-01").(pb.GateClient)
	// rsp, err := target.SayHello(context.Background(), &pb.RequestGate{Name: jglobal.SERVER})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// jlog.Debug(rsp.Message)
}

func (svr *Work) SayHello(ctx context.Context, req *pb.RequestGame) (*pb.GameResponse, error) {
	return &pb.GameResponse{
		Message: "hello, " + req.GetName(),
	}, nil
}
