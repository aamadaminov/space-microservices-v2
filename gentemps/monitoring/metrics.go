package monitoring

import (
metrics "github.com/aamadaminov/space-microservices-v2/pkg/metrics"
configMetrics "github.com/aamadaminov/space-microservices-v2/gentemps/config/metrics"
"fmt"
)

func SetupMetrics(cfg configMetrics.Config) error {
    // if !cfg.Enabled {
    //     return nil
    // }

	// run Exporter for Prometheus
	metrics.SetupPrometheusExporter()

	fmt.Println("ADDRESS_METRICS=", cfg.AddressMetrics)
	go metrics.ServeMetrics(cfg.AddressMetrics)
    
    return nil
}

	
