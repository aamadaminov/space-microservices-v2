package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/aamadaminov/space-microservices/spaceship-omicron/genflightbook/flightbookproto"

	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/opentel"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

type flightbookServiceServer struct {
	pb.UnimplementedFlightbookServiceServer
}

func textGenerator1() string {

	file, err := os.Open("dict8.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	text := ""
	fmt.Println(text)
	scanner := bufio.NewScanner(file)

	for i := 0; i < 5; i++ {
		for j := 0; j < rand.IntN(1_000_000); j++ {
			scanner.Scan()
		}
		text += " " + scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return text
}

func textGenerator2() string {

	file, err := os.Open("dict8.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	text := ""
	fmt.Println(text)
	scanner := bufio.NewScanner(file)

	for i := 0; i < 20; i++ {
		for j := 0; j < rand.IntN(1_000_000); j++ {
			scanner.Scan()
		}
		text += " " + scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return text
}

func (s *flightbookServiceServer) GetFlightbook(ctx context.Context, req *pb.FlightbookRequest) (*pb.FlightbookResponse, error) {
	log.Printf("Request flightbook notes")
	return &pb.FlightbookResponse{T: time.Now().Format("2006-01-02 15:04:05.000"), Text1: textGenerator1(), Text2: textGenerator2()}, nil
}

func main() {
	fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50070)")
	fmt.Println()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Flightbook Generator")

	// setting Otel Exporter Endpoint
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// setting Otel Service Name Endpoint
	otelServiceName := "GenFlightbook"

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
		addressGrpc = ":50071"
	}
	fmt.Println("ADDRESS_GRPC=", addressGrpc)
	listener, err := net.Listen("tcp", addressGrpc)
	if err != nil {
		log.Fatalf("Error starting flightbook generator server: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterFlightbookServiceServer(grpcServer, &flightbookServiceServer{})
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
