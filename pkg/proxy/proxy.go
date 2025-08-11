package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/rancher-centralized-monitoring/pkg/config"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/logging"
)

var logger = logging.SetupLogging()

// BuildServiceProxyURL constructs a Rancher service proxy URL
func BuildServiceProxyURL(namespace, service, port string) string {
	return fmt.Sprintf("%s/k8s/clusters/%s/api/v1/namespaces/%s/services/%s:%s/proxy/",
		config.CFG.RancherApiEndpoint,
		config.CFG.ClusterId,
		namespace,
		service,
		port,
	)
}

// BuildPrometheusURL returns the Prometheus service proxy URL
func BuildPrometheusURL() string {
	return BuildServiceProxyURL(
		config.CFG.PrometheusNamespace,
		config.CFG.PrometheusService,
		config.CFG.PrometheusPort,
	)
}

// BuildLokiURL returns the Loki service proxy URL
func BuildLokiURL() string {
	return BuildServiceProxyURL(
		config.CFG.LokiNamespace,
		config.CFG.LokiService,
		config.CFG.LokiPort,
	)
}

// TestServiceConnectivity tests if a service is reachable via Rancher proxy
func TestServiceConnectivity(serviceURL, serviceName string) error {
	client := &http.Client{Timeout: 10 * time.Second}

	// For Loki, test the /ready endpoint
	testURL := serviceURL
	if serviceName == "loki" {
		testURL += "ready"
	} else if serviceName == "prometheus" {
		testURL += "-/ready"
	}
	// For echo-test, just use the root path (no health endpoint needed)

	logger.Printf("Testing connectivity to %s at: %s", serviceName, testURL)

	req, err := http.NewRequest("GET", testURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(config.CFG.RancherApiAccessKey, config.CFG.RancherApiSecretKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error connecting to %s: %v", serviceName, err)
	}
	defer resp.Body.Close()

	logger.Printf("%s responded with status: %d", serviceName, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s health check failed, status code: %d", serviceName, resp.StatusCode)
	}

	logger.Printf("Successfully connected to %s service", serviceName)
	return nil
}
