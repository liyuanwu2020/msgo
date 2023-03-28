package rpc

import (
	"google.golang.org/grpc"
	"net"
)

//conn, err := grpc.Dial("localhost:9111", grpc.WithTransportCredentials(insecure.NewCredentials()))
//if err != nil {
//panic(err)
//}
//defer func(conn *grpc.ClientConn) {
//	err := conn.Close()
//	if err != nil {
//
//	}
//}(conn)
//rpcClient := api.NewGoodsApiClient(conn)
//rsp, err := rpcClient.Find(context.TODO(), &api.GoodsRequest{})
//if err != nil {
//panic(err)
//}

type MsGrpcServer struct {
	listen   net.Listener
	g        *grpc.Server
	register []func(g *grpc.Server)
	opt      []grpc.ServerOption
}

type MsGrpcOption interface {
	Apply(s *MsGrpcServer)
}

func NewGrpcServer(addr string) (*MsGrpcServer, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	ms := &MsGrpcServer{}
	ms.listen = listen
	ms.g = grpc.NewServer()
	return ms, nil
}
