package telemetry

import (
	"context"
	"log"

	otel "github.com/aamadaminov/space-microservices-v2/pkg/otel"
	configOtel "github.com/aamadaminov/space-microservices-v2/producercoords/config/otel"
)

func SetupOTEL(cfg configOtel.Config) error {
	// if !cfg.Enabled {
	//     return nil
	// }

	// init Otel
	_, err := otel.SetupOpenTelemetry(context.Background(), cfg.OtelExporterEndpoint, cfg.OtelServiceName)
	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry: %v", err)
		//return
	}

	return nil
}
