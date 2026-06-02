package grpc

import (
	configGRPC "github.com/aamadaminov/space-microservices-v2/gencoords/config/grpc"
	pb "github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gencoords/v1/gencoordsv1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"
)

func SetupGRPC(cfg configGRPC.Config) error {
    // if !cfg.Enabled {
    //     return nil
    // }

	fmt.Println("ADDRESS_GRPC=", cfg.AddressGrpc)
	listener, err := net.Listen("tcp", cfg.AddressGrpc)
	if err != nil {
		log.Fatalf("Error starting coords generator server: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	
	pb.RegisterSensorServiceServer(grpcServer, &sensorServiceServer{})
	log.Printf("gRPC server listening at %v\n", listener.Addr())

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
    
    return nil
}

// gRPC settings
type sensorServiceServer struct {
	pb.UnimplementedSensorServiceServer
}

func (s *sensorServiceServer) GetSensor(ctx context.Context, req *pb.SensorRequest) (*pb.SensorResponse, error) {
	log.Printf("Request coords")
	return &pb.SensorResponse{Time: time.Now().Format("2006-01-02 15:04:05.000000"), X1: rand.Float64(), Y1: rand.Float64(), Z1: rand.Float64(), Vx1: rand.Float64(), Vy1: rand.Float64(), Vz1: rand.Float64()}, nil
}	