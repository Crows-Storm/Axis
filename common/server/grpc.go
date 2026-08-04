package server

import (
	"net"

	"github.com/Crows-Storm/Axis/common/config"
	"github.com/Crows-Storm/Axis/common/config/logger"

	grpc_logger "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
)

func RunGRPCServer(registerServer func(server *grpc.Server)) {
	addr := config.Get().GetGRPCAddr()
	if addr == "" {
		logger.Panic("[server.RunGRPCServer] missing GRPCServer configuration.")
	}
	RunGRPCServerOnAddr(addr, registerServer)
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	loggerEntry := logger.Entry()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_tags.UnaryServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.UnaryServerInterceptor(loggerEntry),
		),
		grpc.ChainStreamInterceptor(
			grpc_tags.StreamServerInterceptor(grpc_tags.WithFieldExtractor(grpc_tags.CodeGenRequestFieldExtractor)),
			grpc_logger.StreamServerInterceptor(loggerEntry),
		),
	)
	registerServer(grpcServer)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Panic(err)
	}

	logger.Infof("Starting gRPC Server, Listening: %s", addr)
	if err := grpcServer.Serve(listen); err != nil {
		logger.Panic(err)
	}
}
