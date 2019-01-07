package server

import (
	"andre/tauros/api"
	"context"
	"log"
)

// Ping just returns "Pong".
func (s *TaurosServer) Ping(ctx context.Context, req *api.PingReq) (*api.PingResp, error) {
	log.Printf("Received: %v", req.Msg)
	return &api.PingResp{Msg: "Pong"}, nil
}
