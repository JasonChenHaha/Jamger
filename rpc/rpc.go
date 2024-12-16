package jrpc

var Rpc *rpc

type rpc struct {
}

func Init() {
	Rpc = &rpc{}
	// if jglobal.GROUP == "game" {
	// 	jetcd.WatchJoin("gate", func(group string, server string, info string) {
	// 		jlog.Debug("gate join! ", server, info)
	// 	})
	// 	jetcd.WatchLeave("gate", func(group string, server string, info string) {
	// 		jlog.Debug("gate leave! ", server, info)
	// 	})
	// }
}

// lis, err := net.Listen("tcp", jconfig.GetString("grpc.addr"))
// if err != nil {
// 	log.Fatal(err)
// }

// s := grpc.NewServer()
// pb.RegisterGreeterServer(s, &server{})
// err = s.Serve(lis)
// if err != nil {
// 	log.Fatal(err)
// }
