package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/opentel"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"
	pb "github.com/aamadaminov/space-microservices/spaceship-omicron/producerflightbook/flightbookproto"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50071), ADDRESS_RABBITMQ (default localhost:5672)")
	fmt.Println()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Flightbook Producer")
	fmt.Println()

	// setting Otel Exporter Endpoint
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// setting Otel Service Name Endpoint
	otelServiceName := "FlightbookProducer"

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
		addressMetrics = ":2224"
	}
	fmt.Println("ADDRESS_METRICS=", addressMetrics)
	go prom.ServeMetrics(addressMetrics)

	// start gRPC Client
	addressGrpc := os.Getenv("ADDRESS_GRPC")
	if addressGrpc == "" {
		addressGrpc = "127.0.0.1:50071"
	}
	fmt.Println("ADDRESS_GRPC=", addressGrpc)
	conn, err := grpc.NewClient(addressGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("grpc.NewClient(%q): %v", addressGrpc, err)
	}
	defer conn.Close()
	client := pb.NewFlightbookServiceClient(conn)

	// RabbitMQ start
	addressRabbitMq := os.Getenv("ADDRESS_RABBITMQ")
	if addressRabbitMq == "" {
		addressRabbitMq = "localhost:5672"
	}
	fmt.Println("ADDRESS_RABBITMQ=", addressRabbitMq)
	// RabbitMQ start
	addressRabbitMqStr := fmt.Sprintf("amqp://guest:guest@%s/", addressRabbitMq)
	connRabbit, err := amqp.Dial(addressRabbitMqStr)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connRabbit.Close()

	ch, err := connRabbit.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"flightbook_notes", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for {
		// read coords from coords generator
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		//defer cancel()
		res, err := client.GetFlightbook(ctx, &pb.FlightbookRequest{})
		if err != nil {
			log.Fatalf("Error call GetFlightbook: %v", err)
		}

		body := fmt.Sprintf("%s;%s;%s;", res.T, res.Text1, res.Text2)

		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: res.T,
				Body:          []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		fmt.Printf("[x] Sent %s\n", body)

		time.Sleep(5 * time.Second)
	}

}
