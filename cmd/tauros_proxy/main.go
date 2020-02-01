package main

import (
	"os"
	"andre/tauros/api"
	"andre/tauros/internal/conf"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	if (len(os.Args) != 2) {
		log.Fatal("You shall pass the target hostname:port.")
	}
	target := os.Args[1]
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := api.NewTaurosClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Ping(ctx, &api.PingReq{Msg: "Ping"})
	if err != nil {
		log.Fatalf("could not ping: %v", err)
	}
	log.Printf("Ping msg: %s", r.Msg)
}
