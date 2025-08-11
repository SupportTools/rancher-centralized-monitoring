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

	// Prometheus configuration
	PrometheusNamespace string
	PrometheusService   string
	PrometheusPort      string

	// Loki configuration
	LokiNamespace string
	LokiService   string
	LokiPort      string

	// Generic remote endpoint configuration
	RemoteNamespace string
	RemoteService   string
	RemotePort      string
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
		ClusterName:         getEnvOrDefault("CLUSTER_NAME", ""),

		// Prometheus configuration
		PrometheusNamespace: getEnvOrDefault("PROMETHEUS_NAMESPACE", "cattle-monitoring-system"),
		PrometheusService:   getEnvOrDefault("PROMETHEUS_SERVICE", "rancher-monitoring-prometheus"),
		PrometheusPort:      getEnvOrDefault("PROMETHEUS_PORT", "9090"),

		// Loki configuration
		LokiNamespace: getEnvOrDefault("LOKI_NAMESPACE", "cattle-logging-system"),
		LokiService:   getEnvOrDefault("LOKI_SERVICE", "rancher-logging-loki"),
		LokiPort:      getEnvOrDefault("LOKI_PORT", "3100"),

		// Generic remote endpoint configuration
		RemoteNamespace: getEnvOrDefault("REMOTE_NAMESPACE", ""),
		RemoteService:   getEnvOrDefault("REMOTE_SERVICE", ""),
		RemotePort:      getEnvOrDefault("REMOTE_PORT", ""),
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


func parseEnvBool(key string) bool {
	value := os.Getenv(key)
	boolValue := false
	if value == "true" {
		boolValue = true
	}
	return boolValue
}
