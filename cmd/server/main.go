package main

import (
	"fmt"
	"gRPCService/internal/service"
	pb "gRPCService/internal/service/gRPC_service"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

const (
	storagePath = "./storage"
)

func main() {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	fileServiceServer := service.NewFileServiceServer()
	pb.RegisterFileServiceServer(s, fileServiceServer)

	fmt.Println("Server is running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
