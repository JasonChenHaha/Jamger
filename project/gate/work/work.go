package jwork

import (
	"context"
	"jconfig"
	pb "jpb"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	pb.GreeterServer
}

// ------------------------- outside -------------------------

func Init() {
	lis, err := net.Listen("tcp", jconfig.GetString("grpc.addr"))
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	err = s.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

func (svr *server) SayHello(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	return &pb.Response{
		Message: "hello, " + req.GetName(),
	}, nil
}
