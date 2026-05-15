package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	matchingv1 "github.com/ganesh/papertrading/services/go/matching/pb/papertrading/matching/v1"
)

// server is a placeholder until per-symbol actors land (GitHub #11).
type server struct {
	matchingv1.UnimplementedMatchingServiceServer
}

func main() {
	addr := ":50051"
	if v := os.Getenv("MATCHING_GRPC_ADDR"); v != "" {
		addr = v
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	s := grpc.NewServer()
	matchingv1.RegisterMatchingServiceServer(s, &server{})
	fmt.Printf("matching: gRPC %s (stub)\n", lis.Addr().String())
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
