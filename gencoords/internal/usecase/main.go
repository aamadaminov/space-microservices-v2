package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gencoords/v1/gencoordsv1"
	//"github.com/aamadaminov/space-microservices-v2/pkg/otel"


	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"github.com/aamadaminov/space-microservices-v2/gencoords/config"
	"github.com/aamadaminov/space-microservices-v2/gencoords/monitoring"
	"github.com/aamadaminov/space-microservices-v2/gencoords/telemetry"
)

// gRPC settings
type sensorServiceServer struct {
	pb.UnimplementedSensorServiceServer
}

func (s *sensorServiceServer) GetSensor(ctx context.Context, req *pb.SensorRequest) (*pb.SensorResponse, error) {
	log.Printf("Request coords")
	return &pb.SensorResponse{Time: time.Now().Format("2006-01-02 15:04:05.000000"), X1: rand.Float64(), Y1: rand.Float64(), Z1: rand.Float64(), Vx1: rand.Float64(), Vy1: rand.Float64(), Vz1: rand.Float64()}, nil
}

func main() {

    cfg, err := config.Init()
    if err != nil {
        log.Fatal(err)
    }

    if err := telemetry.SetupOTEL(cfg.OTEL); err != nil {
        log.Fatal(err)
    }
    
    if err := monitoring.SetupMetrics(cfg.Metrics); err != nil {
        log.Fatal(err)
    }
 
	// fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50070)")
	// fmt.Println()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Coords Generator")

	// // setting Otel Exporter Endpoint
	// otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	// if otelExporterEndpoint == "" {
	// 	otelExporterEndpoint = "127.0.0.1:4317"
	// }
	// fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// // setting Otel Service Name Endpoint
	// otelServiceName := "CoordsGen"



	// start gRPC Server
	addressGrpc := os.Getenv("ADDRESS_GRPC")
	if addressGrpc == "" {
		addressGrpc = ":50070"
	}
	fmt.Println("ADDRESS_GRPC=", addressGrpc)
	listener, err := net.Listen("tcp", addressGrpc)
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

	// listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
}
