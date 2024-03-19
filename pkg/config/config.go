package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	Debug               bool
	MetricsPort         string
	RancherApiEndpoint  string
	RancherApiAccessKey string
	RancherApiSecretKey string
	ClusterId           string
	ClusterName         string
	MonitoringNamespace string
	MonitoringService   string
	MonitoringPort      string
}

var CFG Config

func LoadConfigFromEnv() Config {
	config := Config{
		Debug:               parseEnvBool("DEBUG"),
		MetricsPort:         getEnvOrDefault("METRICS_PORT", "9000"),
		RancherApiEndpoint:  getEnvOrDefault("RANCHER_API_ENDPOINT", ""),
		RancherApiAccessKey: getEnvOrDefault("RANCHER_API_ACCESS_KEY", ""),
		RancherApiSecretKey: getEnvOrDefault("RANCHER_API_SECRET_KEY", ""),
		ClusterId:           getEnvOrDefault("CLUSTER_ID", ""),
	}

	CFG = config

	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil {
		log.Printf("Failed to parse environment variable %s: %v. Using default value: %d", key, err, defaultValue)
		return defaultValue
	}
	return intValue
}

func parseEnvBool(key string) bool {
	value := os.Getenv(key)
	boolValue := false
	if value == "true" {
		boolValue = true
	}
	return boolValue
}
