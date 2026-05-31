package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/aamadaminov/space-microservices/spaceship-omicron/gentemps/tempsproto"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/opentel"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

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

func main() {
	fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50069)")
	fmt.Println()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Temps-Humidity Generator")

	// setting Otel Exporter Endpoint
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// setting Otel Service Name Endpoint
	otelServiceName := "TempsGen"

	// init Otel
	_, err := opentel.SetupOpenTelemetry(context.Background(), otelExporterEndpoint, otelServiceName)
	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry: %v", err)
		return
	}

	// run Exporter for Prometheus
	prom.SetupPrometheusExporter()
	addressMetrics := os.Getenv("ADDRESS_METRICS")
	if addressMetrics == "" {
		addressMetrics = ":2223"
	}
	fmt.Println("ADDRESS_METRICS=", addressMetrics)
	go prom.ServeMetrics(addressMetrics)

	// start gRPC Server
	addressGrpc := os.Getenv("ADDRESS_GRPC")
	if addressGrpc == "" {
		addressGrpc = ":50069"
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
