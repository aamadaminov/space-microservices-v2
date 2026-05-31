package metrics

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

// Prometheus exporter
func SetupPrometheusExporter() (*prometheus.Exporter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)
	return exporter, nil
}

func ServeMetrics(address string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
	log.Printf("serving metrics at %s/metrics", address)
}
