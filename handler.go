package main

import (
	"context"
	"github.com/pkg/errors"
	pb "github.com/tsuki42/shippy-service-consignment/proto/consignment"
	vesselProto "github.com/tsuki42/shippy-service-vessel/proto/vessel"
)

type handler struct {
	repository
	vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment -
func (s *handler) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	vesselResponse, err := s.vesselClient.FindAvailable(ctx, &vesselProto.Specification{
		Capacity:  int32(len(req.Containers)),
		MaxWeight: req.Weight,
	})
	if vesselResponse == nil {
		return errors.New("error fetching vessel, returned nil")
	}
	if err != nil {
		return err
	}

	req.VesselId = vesselResponse.Vessel.Id

	if err = s.repository.Create(ctx, MarshalConsignment(req)); err != nil {
		return err
	}
	res.Created = true
	res.Consignment = req
	return nil
}

// GetConsignments -
func (s *handler) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments, err := s.repository.GetAll(ctx)
	if err != nil {
		return err
	}
	res.Consignments = UnmarshalConsignmentCollection(consignments)
	return nil
}
