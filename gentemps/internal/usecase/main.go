package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
	"github.com/aamadaminov/space-microservices-v2/gentemps/config"
	"github.com/aamadaminov/space-microservices-v2/gentemps/monitoring"
	"github.com/aamadaminov/space-microservices-v2/gentemps/telemetry"
	"github.com/aamadaminov/space-microservices-v2/gentemps/controller/grpc"
)

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

	if err := grpc.SetupGRPC(cfg.GRPC); err != nil {
        log.Fatal(err)
    }
 
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("Starting Temps Generator")

	// listen for termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	<-signalCh
}
