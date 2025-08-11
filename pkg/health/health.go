package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/supporttools/rancher-centralized-monitoring/pkg/config"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/logging"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/proxy"
)

// VersionInfo represents the structure of version information.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
}

var logger = logging.SetupLogging()

// version holds the application version. It's set during the build process.
var version = "MISSING VERSION INFO"

// GitCommit holds the Git commit hash of the build. It's set during the build process.
var GitCommit = "MISSING GIT COMMIT"

// BuildTime holds the timestamp of when the build was created. It's set during the build process.
var BuildTime = "MISSING BUILD TIME"

// HealthzHandler returns an HTTP handler function that checks Rancher API connectivity.
func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("HealthzHandler")

		// Test basic Rancher API connectivity
		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("GET", config.CFG.RancherApiEndpoint, nil)
		if err != nil {
			logger.Printf("HealthzHandler: Failed to create request: %v", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		req.SetBasicAuth(config.CFG.RancherApiAccessKey, config.CFG.RancherApiSecretKey)

		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("HealthzHandler: Failed to connect to Rancher API: %v", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Printf("HealthzHandler: Rancher API returned status: %d", resp.StatusCode)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		logger.Printf("HealthzHandler: Rancher API is reachable")
		fmt.Fprintf(w, "ok")
	}
}

// ReadyzHandler returns an HTTP handler function that checks service connectivity via proxy.
func ReadyzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("ReadyzHandler")

		// Test connectivity to configured remote services
		allHealthy := true

		// Test Loki if configured
		if config.CFG.LokiNamespace != "" && config.CFG.LokiService != "" {
			lokiURL := proxy.BuildLokiURL()
			if err := proxy.TestServiceConnectivity(lokiURL, "loki"); err != nil {
				logger.Printf("ReadyzHandler: Loki service check failed: %v", err)
				allHealthy = false
			}
		}

		// Test Prometheus if configured
		if config.CFG.PrometheusNamespace != "" && config.CFG.PrometheusService != "" {
			prometheusURL := proxy.BuildPrometheusURL()
			if err := proxy.TestServiceConnectivity(prometheusURL, "prometheus"); err != nil {
				logger.Printf("ReadyzHandler: Prometheus service check failed: %v", err)
				allHealthy = false
			}
		}

		// Test remote service if configured
		if config.CFG.RemoteNamespace != "" && config.CFG.RemoteService != "" && config.CFG.RemotePort != "" {
			remoteURL := proxy.BuildServiceProxyURL(config.CFG.RemoteNamespace, config.CFG.RemoteService, config.CFG.RemotePort)
			if err := proxy.TestServiceConnectivity(remoteURL, config.CFG.RemoteService); err != nil {
				logger.Printf("ReadyzHandler: Remote service check failed: %v", err)
				allHealthy = false
			}
		}

		if !allHealthy {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}

		logger.Printf("ReadyzHandler: All configured services are reachable")
		fmt.Fprintf(w, "ok")
	}
}

// VersionHandler returns version information as JSON.
func VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("VersionHandler")

		versionInfo := VersionInfo{
			Version:   version,
			GitCommit: GitCommit,
			BuildTime: BuildTime,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(versionInfo); err != nil {
			logger.Printf("Failed to encode version info to JSON: %v", err)
			http.Error(w, "Failed to encode version info", http.StatusInternalServerError)
		}
	}
}