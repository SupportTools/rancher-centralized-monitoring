package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.CFG.RancherInsecureSkipVerify,
			},
		},
	}

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

// createProxyHandler creates an HTTP handler that proxies requests to the specified service URL
func createProxyHandler(serviceURL, serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build target URL by combining service URL with the request path
		targetURL := strings.TrimSuffix(serviceURL, "/") + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		logger.Printf("Proxying %s request to %s: %s", serviceName, r.Method, targetURL)

		// Create the proxy request
		proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			logger.Printf("Error creating proxy request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Copy headers from original request
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		// Set Rancher authentication
		proxyReq.SetBasicAuth(config.CFG.RancherApiAccessKey, config.CFG.RancherApiSecretKey)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.CFG.RancherInsecureSkipVerify,
				},
			},
		}

		// Execute the proxy request
		resp, err := client.Do(proxyReq)
		if err != nil {
			logger.Printf("Error executing proxy request to %s: %v", serviceName, err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Set response status
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			logger.Printf("Error copying response body from %s: %v", serviceName, err)
		}
	}
}

// PrometheusHandler returns an HTTP handler for proxying requests to Prometheus
func PrometheusHandler() http.HandlerFunc {
	prometheusURL := BuildPrometheusURL()
	return createProxyHandler(prometheusURL, "prometheus")
}

// LokiHandler returns an HTTP handler for proxying requests to Loki
func LokiHandler() http.HandlerFunc {
	lokiURL := BuildLokiURL()
	return createProxyHandler(lokiURL, "loki")
}

// RemoteServiceHandler returns an HTTP handler for proxying requests to a custom remote service
func RemoteServiceHandler() http.HandlerFunc {
	if config.CFG.RemoteNamespace == "" || config.CFG.RemoteService == "" || config.CFG.RemotePort == "" {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Remote service not configured", http.StatusServiceUnavailable)
		}
	}

	remoteURL := BuildServiceProxyURL(config.CFG.RemoteNamespace, config.CFG.RemoteService, config.CFG.RemotePort)
	return createProxyHandler(remoteURL, config.CFG.RemoteService)
}
