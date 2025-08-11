package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/rancher-centralized-monitoring/pkg/logging"
)

var logger = logging.SetupLogging()
var startTime = time.Now()

// MetricsHandler returns a simple metrics endpoint handler
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("MetricsHandler")

		// Basic metrics - can be extended with actual Prometheus metrics later
		metrics := `# HELP rancher_monitoring_relay_info Information about the Rancher monitoring relay
# TYPE rancher_monitoring_relay_info gauge
rancher_monitoring_relay_info{version="0.1.0"} 1

# HELP rancher_monitoring_relay_requests_total Total number of HTTP requests handled
# TYPE rancher_monitoring_relay_requests_total counter
rancher_monitoring_relay_requests_total{endpoint="/metrics"} 1

# HELP rancher_monitoring_relay_uptime_seconds Uptime of the service in seconds
# TYPE rancher_monitoring_relay_uptime_seconds gauge
rancher_monitoring_relay_uptime_seconds ` + formatFloat(time.Since(startTime).Seconds()) + `
`

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(metrics)); err != nil {
			logger.Printf("Error writing metrics response: %v", err)
		}
	}
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
