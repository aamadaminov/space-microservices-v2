package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aamadaminov/space-microservices/spaceship-omicron/consumernotes/postgresqlred"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func initTracer() func(context.Context) error {
	ctx := context.Background()
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otelExporterEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// setting Otel Service Name Endpoint
	otelServiceName := "ConsumerNotes"

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(otelServiceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp.Shutdown
}

func main() {
	fmt.Println(`List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50071),
	 ADDRESS_RABBITMQ (default localhost:5672), API_ADDRESS (default :9071), ADDRESS_POSTGRESQL (localhost:5432), ADDRESS_REDIS (localhost:6379)`)
	fmt.Println()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting All Notes Consumer")

	// Init tracer + server listening
	apiAddress := os.Getenv("API_ADDRESS")
	if apiAddress == "" {
		apiAddress = ":9071"
	}
	fmt.Println("API_ADDRESS=", apiAddress)

	shutdown := initTracer()
	defer shutdown(context.Background())

	go func() {
		// change speed request handler
		http.HandleFunc("/api/", postgresqlred.GetJournalHandler)
		err5 := http.ListenAndServe(apiAddress, nil)
		if err5 != nil {
			log.Println("error starting server:", err5)
		}
	}()

	// run Exporter for Prometheus
	prom.SetupPrometheusExporter()
	addressMetrics := os.Getenv("ADDRESS_METRICS")
	if addressMetrics == "" {
		addressMetrics = ":2225"
	}
	fmt.Println("ADDRESS_METRICS=", addressMetrics)
	go prom.ServeMetrics(addressMetrics)

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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	//ctx := context.Background()
	dbRes, err := postgresqlred.NewApp(context.Background())

	if err != nil {
		log.Println("link fail")
	}
	//fmt.Println(dbRes)

	go func() {
		for d := range msgs {
			fmt.Printf("Received a message: %s\n", d.Body)

			// Save to PostgreSQL
			ctx := context.Background()
			postgresqlred.CreateNotesHandler(dbRes, d.Body, ctx)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	// listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh

}
