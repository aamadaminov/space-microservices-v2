package grpc

import (
	configGRPC "github.com/aamadaminov/space-microservices-v2/gentemps/config/grpc"
	pb "github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gentemps/v1/gentempsv1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"math"
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
		log.Fatalf("Error starting temps generator server: %v", err)
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

func tempRand() float64 {
	minFloat := 18.5
	maxFloat := 25.7
	tempRand := math.Round(100*(minFloat+rand.Float64()*(maxFloat-minFloat))) / 100
	return tempRand
}

func humidityRand() float64 {
	minFloat := 35.0
	maxFloat := 62.0
	humidityRand := math.Round(100*(minFloat+rand.Float64()*(maxFloat-minFloat))) / 100
	return humidityRand
}

func (s *sensorServiceServer) GetSensor(ctx context.Context, req *pb.SensorRequest) (*pb.SensorResponse, error) {
	log.Printf("Request temps")

	return &pb.SensorResponse{Time: time.Now().Format("2006-01-02 15:04:05.000000"), T1: tempRand(), H1: humidityRand(), T2: tempRand(), H2: humidityRand(), T3: tempRand(), H3: humidityRand(), T4: tempRand(), H4: humidityRand(), T5: tempRand(), H5: humidityRand(), T6: tempRand(), H6: humidityRand()}, nil
}
