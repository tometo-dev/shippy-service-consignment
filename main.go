package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/tsuki42/shippy-service-consignment/proto/consignment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository -- dummy repository. Will be updated later.
type Repository struct {
	mu           sync.RWMutex
	consignments []*pb.Consignment
}

// Create a new consignment
func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	repo.mu.Lock()
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	repo.mu.Unlock()
	return consignment, nil
}

// GetAll consignments
func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

// Service implements all the methods to satisfy the service
type service struct {
	repo repository
}

// CreateConsignment method
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {
	// Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return nil, err
	}

	return &pb.Response{Created: true, Consignment: consignment}, nil
}

// GetConsignments -
func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {
	consignments := s.repo.GetAll()
	return &pb.Response{Consignments: consignments}, nil
}

func main() {
	repo := &Repository{}

	// Setup gRPC server
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// Register our serice with gRPC server
	pb.RegisterShippingServiceServer(s, &service{repo})

	// Register reflection service on gRPC server
	reflection.Register(s)

	log.Println("Running on port: ", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
