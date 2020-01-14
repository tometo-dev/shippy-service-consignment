package main

import (
	"context"
	"fmt"
	"github.com/micro/go-micro"
	pb "github.com/tsuki42/shippy-service-consignment/proto/consignment"
	vesselProto "github.com/tsuki42/shippy-service-vessel/proto/vessel"
	"log"
	"os"
)

const (
	defaultHost = "datastore:27017"
)

func main() {
	// Create a new service.
	srv := micro.NewService(
		// This name must match the package name given in the protobuf definition
		micro.Name("shippy.service.consignment"),
	)
	// Init will parse the command line flags
	srv.Init()

	uri := os.Getenv("DB_HOST")
	if uri == "" {
		uri = defaultHost
	}

	client, err := CreateClient(context.Background(), uri, 0)
	if err != nil {
		log.Panic(err)
	}
	defer client.Disconnect(context.Background())

	consignmentCollection := client.Database("shippy").Collection("consignments")

	repository := &MongoRepository{consignmentCollection}

	vesselClient := vesselProto.NewVesselServiceClient("shippy.service.vessel", srv.Client())

	h := &handler{repository, vesselClient}

	// Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), h)

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
