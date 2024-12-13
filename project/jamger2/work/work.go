package jwork

// ------------------------- outside -------------------------

func Init() {
	// con, err := grpc.NewClient(jconfig.GetString("grpc.addr"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	jlog.Fatal(err)
	// }
	// c := pb.NewGreeterClient(con)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// rsp, err := c.SayHello(ctx, &pb.Request{Name: "jamger2"})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// jlog.Debug(rsp.Message)
}
