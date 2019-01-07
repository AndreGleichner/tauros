package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"andre/tauros/api"
	"google.golang.org/grpc/reflection"
	"andre/tauros/internal/server"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)
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