package service

import (
	"context"
	pb "gRPCService/internal/service/gRPC_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"path/filepath"
	"time"
)

const (
	maxUploadConcurrency = 10
	maxListConcurrency   = 100
	storagePath          = "./storage"
)

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	uploadSemaphore chan struct{}
	listSemaphore   chan struct{}
}

func NewFileServiceServer() *FileServiceServer {
	return &FileServiceServer{
		uploadSemaphore: make(chan struct{}, maxUploadConcurrency),
		listSemaphore:   make(chan struct{}, maxListConcurrency),
	}
}

func (s *FileServiceServer) GetFiles(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	s.uploadSemaphore <- struct{}{}
	defer func() { <-s.uploadSemaphore }()

	filePath := filepath.Join(storagePath, req.FileName)
	err := os.WriteFile(filePath, req.Data, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save file: %v", err)
	}
	return &pb.GetFileResponse{Success: true, Message: "File uploaded successfully"}, nil
}

func (s *FileServiceServer) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	s.uploadSemaphore <- struct{}{}
	defer func() { <-s.uploadSemaphore }()

	filePath := filepath.Join(storagePath, req.FileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return status.Errorf(codes.NotFound, "file not found: %v", err)
	}

	if err := stream.Send(&pb.DownloadResponse{Data: data}); err != nil {
		return status.Errorf(codes.Internal, "failed to send file: %v", err)
	}
	return nil
}

func (s *FileServiceServer) ListFiles(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileResponse, error) {
	s.listSemaphore <- struct{}{}
	defer func() { <-s.listSemaphore }()

	files, err := os.ReadDir(storagePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read directory: %v", err)
	}

	var fileInfos []*pb.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get file info: %v", err)
		}
		fileInfos = append(fileInfos, &pb.FileInfo{
			FileName:  file.Name(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
			UpdatedAt: info.ModTime().Format(time.RFC3339),
		})
	}
	return &pb.ListFileResponse{Files: fileInfos}, nil
}
