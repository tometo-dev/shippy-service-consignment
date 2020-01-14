package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro"
	pb "github.com/tsuki42/shippy-service-consignment/proto/consignment"
	vesselProto "github.com/tsuki42/shippy-service-vessel/proto/vessel"
	"log"
	"sync"
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
	repo         repository
	vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment method
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

	// Here we call a client instance of our vessel service with our consignment weight,
	// and the amount of containers as the capacity value
	vesselReponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight, Capacity: int32(len(req.Containers)),
	})
	log.Printf("Found vessel: %s \n", vesselReponse.Vessel.Name)
	if err != nil {
		return err
	}

	// We set the VesselId as the vessel we got back from our vessel service
	req.VesselId = vesselReponse.Vessel.Id

	// Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = consignment

	return nil
}

// GetConsignments -
func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}

func main() {
	repo := &Repository{}

	// Create a new service.
	srv := micro.NewService(
		// This name must match the package name given in the protobuf definition
		micro.Name("shippy.service.consignment"),
	)
	// Init will parse the command line flags
	srv.Init()

	vesselClient := vesselProto.NewVesselServiceClient("shippy.service.vessel", srv.Client())

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
