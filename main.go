package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/rancher-centralized-monitoring/pkg/config"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/health"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/logging"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/metrics"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/proxy"
)

var logger = logging.SetupLogging()

func main() {
	logger.Println("Starting Rancher Centralized Monitoring Agent")
	if config.CFG.Debug {
		logger.Println("Debug mode enabled")
	}

	config.LoadConfigFromEnv()

	// Check if environment variables are set
	if config.CFG.RancherApiEndpoint == "" {
		logger.Fatal("RANCHER_API_ENDPOINT environment variable not set")
	}
	if config.CFG.RancherApiAccessKey == "" {
		logger.Fatal("RANCHER_API_ACCESS_KEY environment variable not set")
	}
	if config.CFG.RancherApiSecretKey == "" {
		logger.Fatal("RANCHER_API_SECRET_KEY environment variable not set")
	}
	if config.CFG.ClusterId == "" {
		logger.Fatal("CLUSTER_ID environment variable not set")
	}

	// Verify access to Rancher API
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", config.CFG.RancherApiEndpoint, nil)
	if err != nil {
		logger.Fatal("Error creating request: ", err)
	}
	req.SetBasicAuth(config.CFG.RancherApiAccessKey, config.CFG.RancherApiSecretKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Fatal("Error connecting to Rancher: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Fatal("Failed to connect to Rancher, status code: ", resp.StatusCode)
	}

	logger.Println("Successfully connected to Rancher API")

	// Test connectivity to Loki service via proxy
	lokiURL := proxy.BuildLokiURL()
	logger.Printf("Testing Loki connectivity at: %s", lokiURL)
	
	if err := proxy.TestServiceConnectivity(lokiURL, "loki"); err != nil {
		logger.Printf("Warning: Failed to connect to Loki service: %v", err)
	}

	// Test connectivity to Prometheus service via proxy if configured
	if config.CFG.PrometheusNamespace != "" {
		prometheusURL := proxy.BuildPrometheusURL()
		logger.Printf("Testing Prometheus connectivity at: %s", prometheusURL)
		
		if err := proxy.TestServiceConnectivity(prometheusURL, "prometheus"); err != nil {
			logger.Printf("Warning: Failed to connect to Prometheus service: %v", err)
		}
	}

	// Test connectivity to custom remote service if configured
	if config.CFG.RemoteNamespace != "" && config.CFG.RemoteService != "" && config.CFG.RemotePort != "" {
		remoteURL := proxy.BuildServiceProxyURL(config.CFG.RemoteNamespace, config.CFG.RemoteService, config.CFG.RemotePort)
		logger.Printf("Testing remote service connectivity at: %s", remoteURL)
		
		if err := proxy.TestServiceConnectivity(remoteURL, config.CFG.RemoteService); err != nil {
			logger.Printf("Warning: Failed to connect to remote service: %v", err)
		}
	}

	// Setup HTTP endpoints
	http.HandleFunc("/health", health.HealthzHandler())
	http.HandleFunc("/ready", health.ReadyzHandler())
	http.HandleFunc("/version", health.VersionHandler())
	http.HandleFunc("/metrics", metrics.MetricsHandler())

	// Start HTTP server with timeouts (security: prevent slowloris attacks)
	address := fmt.Sprintf(":%s", config.CFG.MetricsPort)
	logger.Printf("Starting HTTP server on %s", address)
	
	server := &http.Server{
		Addr:           address,
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	
	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("HTTP server failed to start: %v", err)
	}
}
