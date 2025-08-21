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
	req, err := http.NewRequest("GET", config.CFG.RancherApiEndpoint, http.NoBody)
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

	// Setup metrics/health HTTP server (default port 9000)
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/health", health.HealthzHandler())
	metricsMux.HandleFunc("/ready", health.ReadyzHandler())
	metricsMux.HandleFunc("/version", health.VersionHandler())
	metricsMux.HandleFunc("/metrics", metrics.MetricsHandler())

	metricsAddress := fmt.Sprintf(":%s", config.CFG.MetricsPort)
	logger.Printf("Starting metrics HTTP server on %s", metricsAddress)

	metricsServer := &http.Server{
		Addr:              metricsAddress,
		Handler:           metricsMux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start metrics server in background
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Metrics HTTP server failed to start: %v", err)
		}
	}()

	// Setup Prometheus proxy server on port 9090
	if config.CFG.PrometheusNamespace != "" {
		prometheusMux := http.NewServeMux()
		prometheusMux.HandleFunc("/", proxy.PrometheusHandler())

		prometheusAddress := ":9090"
		logger.Printf("Starting Prometheus proxy server on %s -> %s", prometheusAddress, proxy.BuildPrometheusURL())

		prometheusServer := &http.Server{
			Addr:              prometheusAddress,
			Handler:           prometheusMux,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Start Prometheus proxy server in background
		go func() {
			if err := prometheusServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Prometheus proxy server failed to start: %v", err)
			}
		}()
	}

	// Setup Loki proxy server on port 3100
	lokiMux := http.NewServeMux()
	lokiMux.HandleFunc("/", proxy.LokiHandler())

	lokiAddress := ":3100"
	logger.Printf("Starting Loki proxy server on %s -> %s", lokiAddress, proxy.BuildLokiURL())

	lokiServer := &http.Server{
		Addr:              lokiAddress,
		Handler:           lokiMux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start Loki proxy server in background
	go func() {
		if err := lokiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Loki proxy server failed to start: %v", err)
		}
	}()

	// Setup custom remote service proxy if configured
	if config.CFG.RemoteNamespace != "" && config.CFG.RemoteService != "" && config.CFG.RemotePort != "" {
		remoteMux := http.NewServeMux()
		remoteMux.HandleFunc("/", proxy.RemoteServiceHandler())

		remoteAddress := fmt.Sprintf(":%s", config.CFG.RemotePort)
		logger.Printf("Starting remote service proxy on %s -> %s",
			remoteAddress,
			proxy.BuildServiceProxyURL(config.CFG.RemoteNamespace, config.CFG.RemoteService, config.CFG.RemotePort))

		remoteServer := &http.Server{
			Addr:              remoteAddress,
			Handler:           remoteMux,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		}

		// Start remote service proxy in background
		go func() {
			if err := remoteServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Remote service proxy server failed to start: %v", err)
			}
		}()
	}

	logger.Println("All proxy servers started successfully")

	// Keep the main goroutine alive
	select {}
}
