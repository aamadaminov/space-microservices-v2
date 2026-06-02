package main

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"fmt"

	"log"
	"time"

	pb "github.com/aamadaminov/space-microservices-v2/pkg/gen/proto/gencoords/v1/gencoordsv1"
	"github.com/aamadaminov/space-microservices-v2/producercoords/adapter/kafka"
	"github.com/aamadaminov/space-microservices-v2/producercoords/config"
	"github.com/aamadaminov/space-microservices-v2/producercoords/monitoring"
	"github.com/aamadaminov/space-microservices-v2/producercoords/telemetry"
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

	fmt.Println("ADDRESS_GRPC=", cfg.GRPC.AddressGrpc)
	conn, err := grpc.NewClient(cfg.GRPC.AddressGrpc, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("grpc.NewClient(%q): %v", cfg.GRPC.AddressGrpc, err)
	}
	defer conn.Close()
	client := pb.NewSensorServiceClient(conn)

	// fmt.Println("List of ENVs: OTEL_EXPORTER_OTLP_ENDPOINT (default 127.0.0.1:4317), ADDRESS_METRICS (default :2223), ADDRESS_GRPC (default :50070), ADDRESS_KAFKA (default :9092)")
	// fmt.Println()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Coords Producer")
	fmt.Println()

	go func() {
		// change speed request handler
		http.HandleFunc("/api/", speedHandler)
		err5 := http.ListenAndServe(":9070", nil)
		if err5 != nil {
			log.Println("error starting server:", err5)
		}
	}()

	fmt.Println("ADDRESS_KAFKA=", cfg.Kafka.AddressKafka)

	// init Kafka producer + Otel
	brokers := []string{cfg.Kafka.AddressKafka}
	pNewKafkaProducer, err := kafka.NewKafkaProducer(brokers)
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}
	log.Println("Producer initialized")
	defer pNewKafkaProducer.Close()

	var speedInt time.Duration = 5000

	///// init tracer /////
	//_ = kafka.InitTracer(otelServiceName)

	for {

		// read coords from coords generator
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		res, err := client.GetSensor(ctx, &pb.SensorRequest{})
		if err != nil {
			log.Fatalf("Error call GetSensor: %v", err)
		}
		value := []byte(fmt.Sprintf("%s;%f;%f;%f;%f;%f;%f;", res.Time, res.X1, res.Y1, res.Z1, res.Vx1, res.Vy1, res.Vz1))
		err2 := pNewKafkaProducer.SendMessage("coords-topic", "my-key", value)
		if err != nil {
			log.Fatalf("Failed to send message to Kafka: %v", err2)
		}
		//fmt.Println("Topic offset: ", offset)
		fmt.Printf("Send coords to Kafka: Time: %s, X=%f, Y=%f, Z=%f. Vector: X=%f, Y=%f, Z=%f\n", res.Time, res.X1, res.Y1, res.Z1, res.Vx1, res.Vy1, res.Vz1)

		if setSpeed > 0 {
			speedInt = time.Duration(setSpeed)
			fmt.Println(speedInt)
		}

		time.Sleep(speedInt * time.Millisecond)
	}

}
