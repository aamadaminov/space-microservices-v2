package main

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"fmt"

	"log"
	"time"

	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/opentel"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/pkg/prom"
	"github.com/aamadaminov/space-microservices/spaceship-omicron/producercoords/tracerkafka"
	pb "github.com/aamadaminov/space-microservices/spaceship-omicron/producertemps/tempsproto"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	setSpeed int
)

func speedHandler(w http.ResponseWriter, r *http.Request) {
	nspeedStr := r.URL.Query().Get("speed")
	checkSpeed, err := strconv.Atoi(nspeedStr)
	if err != nil || checkSpeed <= 0 {
		w.Write([]byte("Incorrect speed was set"))
	} else {
		w.Write([]byte("Speed was set:" + nspeedStr))
		setSpeed = checkSpeed
	}
}

func main() {
	fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50069), ADDRESS_KAFKA (default :9092)")
	fmt.Println()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Temps Producer")
	fmt.Println()

	go func() {
		// change speed request handler
		http.HandleFunc("/api/", speedHandler)
		err5 := http.ListenAndServe(":9000", nil)
		if err5 != nil {
			log.Println("error starting server:", err5)
		}
	}()

	// setting Otel Exporter Endpoint
	otelExporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelExporterEndpoint == "" {
		otelExporterEndpoint = "127.0.0.1:4317"
	}
	fmt.Println("OTEL_EXPORTER_OTLP_ENDPOINT=", otelExporterEndpoint)

	// setting Otel Service Name Endpoint
	otelServiceName := "ProducerTemps"

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
		addressGrpc = "127.0.0.1:50069"
	}
	fmt.Println("ADDRESS_GRPC=", addressGrpc)
	conn, err := grpc.NewClient(addressGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("grpc.NewClient(%q): %v", addressGrpc, err)
	}
	defer conn.Close()
	client := pb.NewSensorServiceClient(conn)

	// start Kafka Client
	addressKafka := os.Getenv("ADDRESS_KAFKA")
	if addressKafka == "" {
		addressKafka = ":9092"
	}
	fmt.Println("ADDRESS_KAFKA=", addressKafka)

	// init Kafka producer + Otel
	brokers := []string{addressKafka}
	pNewKafkaProducer, err := tracerkafka.NewKafkaProducer(brokers)
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}
	log.Println("Producer initialized")
	defer pNewKafkaProducer.Close()

	var speedInt time.Duration = 5000

	///// init tracer /////
	//_ = tracerkafka.InitTracer(otelServiceName)

	for {

		// read coords from coords generator
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		res, err := client.GetSensor(ctx, &pb.SensorRequest{})
		if err != nil {
			log.Fatalf("Error call GetSensor: %v", err)
		}
		value := []byte(fmt.Sprintf("%s;%f;%f;%f;%f;%f;%f;%f;%f;%f;%f;%f;%f", res.Time, res.T1, res.H1, res.T2, res.H2, res.T3, res.H3, res.T4, res.H4, res.T5, res.H5, res.T6, res.H6))
		err2 := pNewKafkaProducer.SendMessage("temps-topic", "my-key", value)
		if err != nil {
			log.Fatalf("Failed to send message to Kafka: %v", err2)
		}
		//fmt.Println("Topic offset: ", offset)
		fmt.Printf("Send temps to Kafka: Time: %s, T1=%f, H1=%f, T2=%f, H2=%f,T3=%f, H3=%f,T4=%f, H4=%f,T5=%f, H5=%f, T6=%f, H6=%f\n", res.Time, res.T1, res.H1, res.T2, res.H2, res.T3, res.H3, res.T4, res.H4, res.T5, res.H5, res.T6, res.H6)

		if setSpeed > 0 {
			speedInt = time.Duration(setSpeed)
			fmt.Println(speedInt)
		}

		time.Sleep(speedInt * time.Millisecond)
	}

}
