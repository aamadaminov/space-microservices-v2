package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.38.0"
	"go.uber.org/zap"
)

// KafkaProducer manages Kafka producer operations and maintains connection to Kafka brokers
type KafkaProducer struct {
	producer sarama.SyncProducer // Synchronous producer instance
	brokers  []string            // List of Kafka broker addresses
}

// NewKafkaProducer creates a new KafkaProducer with the specified broker configuration
func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
	// Configure Kafka producer settings
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all replicas to acknowledge
	config.Producer.Retry.Max = 5                    // Retry up to 5 times on failure
	config.Producer.Return.Successes = true          // Required for SyncProducer

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
		brokers:  brokers,
	}, nil
}

func (p *KafkaProducer) SendMessage(topic string, key string, value []byte) error {

	tracer := otel.Tracer("kafka-producer")
	ctx, span := tracer.Start(context.Background(), "produceMessage")
	defer span.End()

	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	var headers []sarama.RecordHeader
	for k, v := range carrier {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Value:   sarama.ByteEncoder(value),
		Headers: headers,
	}

	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		span.RecordError(err)
		zap.L().Error("Failed to send message",
			zap.String("topic", topic),
			zap.Error(err))
		return err
	}

	log.Printf("Message sent successfully: topic %s, partition %d, offset %d\n", topic, partition, offset)

	return nil
}

// Close gracefully shuts down the Kafka producer
func (p *KafkaProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		log.Fatalln("Failed to close producer", err)
		return err
	}
	return nil
}

// --v
func InitTracer(serviceName string) *trace.TracerProvider {

	headers := map[string]string{
		"content-type": "application/json",
	}

	// exporter, err := otlptrace.New(
	// 	context.Background(),
	// 	otlptracehttp.NewClient(
	// 		//--v
	// 		otlptracehttp.WithEndpoint(serviceName),
	// 		otlptracehttp.WithHeaders(headers),
	// 		otlptracehttp.WithInsecure(),
	// 	),
	// )

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			//--v
			otlptracegrpc.WithEndpoint(serviceName),
			otlptracegrpc.WithHeaders(headers),
			otlptracegrpc.WithInsecure(),
		),
	)

	if err != nil {
		zap.L().Fatal("Failed to create stdout exporter", zap.Error(err))
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				//--v
				semconv.ServiceNameKey.String(serviceName),
			)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
