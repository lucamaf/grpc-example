package main

import (
	"log"
	"net"
	"strings"

	pb "../customer"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement customer.CustomerServer.
type server struct {
	savedCustomers []*pb.CustomerRequest
}

// CreateCustomer creates a new Customer
func (s *server) CreateCustomer(ctx context.Context, in *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	s.savedCustomers = append(s.savedCustomers, in)
	//add logging
	log.Printf("Create customer %s", in.Name)

	return &pb.CustomerResponse{Id: in.Id, Success: true}, nil
}

// GetCustomers returns all customers by given filter
func (s *server) GetCustomers(filter *pb.CustomerFilter, stream pb.Customer_GetCustomersServer) error {
	for _, customer := range s.savedCustomers {
		if filter.Keyword != "" {
			if !strings.Contains(customer.Name, filter.Keyword) {
				continue
			}
		}
		if err := stream.Send(customer); err != nil {
			return err
		}
	}

	//add logging
	log.Printf("Get all customers")

	return nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	//setup TLS certificate
	creds, err := credentials.NewServerTLSFromFile("../cert/service.pem", "../cert/service.key")
	if err != nil {
		log.Fatalf("Failed to setup TLS: %v", err)
	}
	// Creates a new gRPC server
	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterCustomerServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	s.Serve(lis)
}
