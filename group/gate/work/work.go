package jwork

import (
	"context"
	"jglobal"
	pb "jpb"
	"jrpc"
)

type Work struct {
	pb.GateServer
}

// ------------------------- outside -------------------------

func Init() {
	jrpc.Server(&pb.Gate_ServiceDesc, &Work{})
	jrpc.Connect(jglobal.SVR_GAME, pb.NewGameClient)
}

func (svr *Work) SayHello(ctx context.Context, req *pb.RequestGate) (*pb.GateResponse, error) {
	return &pb.GateResponse{
		Message: "hello, " + req.GetName(),
	}, nil
}
