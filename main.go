package main

import (
	"net/http"
	"time"

	"github.com/supporttools/rancher-centralized-monitoring/pkg/config"
	"github.com/supporttools/rancher-centralized-monitoring/pkg/logging"
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
}
