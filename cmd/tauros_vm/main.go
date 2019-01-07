package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"andre/tauros/api"
	"andre/tauros/internal/conf"
	"google.golang.org/grpc/reflection"
	"andre/tauros/internal/server"
)

func main() {
	lis, err := net.Listen("tcp", conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	api.RegisterTaurosServer(s, &server.TaurosServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}