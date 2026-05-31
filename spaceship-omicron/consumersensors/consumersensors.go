package main

import (
	"github.com/aamadaminov/space-microservices/spaceship-omicron/consumersensors/clickhousesaver"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/consumersensors/consumer"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/opentel"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// Topic1Handler processes product update events
// func topic1Handler(message *sarama.ConsumerMessage, clickhouseAddress string) error {
func topicHandler1(message *sarama.ConsumerMessage) error {
	// Create context for OpenTelemetry
	tracer := otel.Tracer("kafka-consumer")
	carrier := propagation.MapCarrier{}
	for _, h := range message.Headers {
		carrier[string(h.Key)] = string(h.Value)
	}

	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(context.Background(), carrier)

	// Create a new span and link it to the parent trace ID
	ctx, span := tracer.Start(ctx, "consumeMessageCoords")
	span.SetAttributes(attribute.String("kafka.topic.coords", message.Topic))
	defer span.End()
	fmt.Printf("Value: %s. Topic: %s. Part.: %d. Offset: %d\n", message.Value, message.Topic, message.Partition, message.Offset)

	//clickhousesaver.ClickhouseSaveCoords(ctx, message.Value, message.Offset)

	if message.Topic == "temps-topic" {
		clickhousesaver.ClickhouseSaveTemps(ctx, message.Value, message.Offset)
	}
	if message.Topic == "coords-topic" {
		clickhousesaver.ClickhouseSaveCoords(ctx, message.Value, message.Offset)
	}

	return nil
}

// // Topic1Handler processes product update events
// // func topic1Handler(message *sarama.ConsumerMessage, clickhouseAddress string) error {
// func topicHandler2(message *sarama.ConsumerMessage) error {
// 	// Create context for OpenTelemetry
// 	tracer := otel.Tracer("kafka-consumer-temps")
// 	carrier := propagation.MapCarrier{}
// 	for _, h := range message.Headers {
// 		carrier[string(h.Key)] = string(h.Value)
// 	}

// 	propagator := otel.GetTextMapPropagator()
// 	ctx := propagator.Extract(context.Background(), carrier)

// 	// Create a new span and link it to the parent trace ID
// 	ctx, span := tracer.Start(ctx, "consumeMessageTemps")
// 	span.SetAttributes(attribute.String("kafka.topic.temps", message.Topic))
// 	defer span.End()
// 	fmt.Printf("Value: %s. Topic: %s. Part.: %d. Offset: %d\n", message.Value, message.Topic, message.Partition, message.Offset)

// 	// if message.Topic == "coords-topic" {
// 	// 	clickhousesaver.ClickhouseSaveCoords(ctx, message.Value, message.Offset)
// 	// }

// 	clickhousesaver.ClickhouseSaveTemps(ctx, message.Value, message.Offset)

// 	return nil
// }

func main() {
	fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_KAFKA (default :9092)")
	fmt.Println()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting All Sensors Consumer")

	// setting Otel Exporter Endpoint
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// setting Otel Service Name Endpoint
	otelServiceName := "ConsumerSensors"

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
		addressMetrics = ":2225"
	}
	fmt.Println("ADDRESS_METRICS=", addressMetrics)
	go prom.ServeMetrics(addressMetrics)

	// Clickhouse connect & db create if not exists
	clickhouseAddress := os.Getenv("CLICKHOUSE_ADDRESS")
	if clickhouseAddress == "" {
		clickhouseAddress = "localhost:9000"
	}
	fmt.Println("CLICKHOUSE_ADDRESS=", clickhouseAddress)
	clickhousesaver.ClickhouseConn(clickhouseAddress)
	clickhousesaver.ClickhouseServerVersion(clickhouseAddress)
	clickhousesaver.ClickhouseCreateTableCoords(clickhouseAddress)
	clickhousesaver.ClickhouseCreateTableTemps(clickhouseAddress)

	// start Kafka Client
	addressKafka := os.Getenv("ADDRESS_KAFKA")
	if addressKafka == "" {
		addressKafka = ":9092"
	}
	fmt.Println("ADDRESS_KAFKA=", addressKafka)
	// Kafka Consumer + Otel
	brokers := []string{addressKafka}
	groupid := "1"
	initOffset := "oldest"
	pNewKafkaConsumer, err := consumer.NewKafkaConsumer(brokers, groupid, initOffset)
	if err != nil {
		log.Fatalln("Failed to create consumer", err)
		os.Exit(1)
	}
	log.Println("Consumer initialized")

	// Register message handlers for Kafka topic
	pNewKafkaConsumer.RegisterHandler("coords-topic", topicHandler1)
	pNewKafkaConsumer.RegisterHandler("temps-topic", topicHandler1)

	// Start consuming messages from Kafka
	if err := pNewKafkaConsumer.Start(); err != nil {
		log.Fatalln("Failed to start consumer", err)
	}
	log.Println("Consumer started successfully")

	// Setup graceful shutdown handling
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	log.Println("Shutting down...")
	pNewKafkaConsumer.Stop()
}
