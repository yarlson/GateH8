package internal

import (
	"encoding/json"
	"os"
	"strings"
)

// APIGateway represents metadata about the API Gateway, including its name and version.
type APIGateway struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CORSConfig encapsulates the configuration for handling Cross-Origin Resource Sharing (CORS)
// at both the vhost and individual endpoint levels.
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowedOrigins"`
	AllowedMethods   []string `json:"allowedMethods"`
	AllowedHeaders   []string `json:"allowedHeaders"`
	ExposedHeaders   []string `json:"exposedHeaders"`
	AllowCredentials bool     `json:"allowCredentials"`
	MaxAge           int      `json:"maxAge"`
}

// Endpoint represents a specific route or API endpoint, detailing its
// path, the supported methods, backend service configuration, and any
// CORS policies specific to this endpoint.
type Endpoint struct {
	CORS    *CORSConfig `json:"cors,omitempty"`
	Path    string      `json:"path"`
	Methods []string    `json:"methods"`
	Backend *Backend    `json:"backend"`
}

// Backend defines the actual service to which the API Gateway will
// route the requests. This includes the service URL and any associated
// timeout settings.
type Backend struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

// GetProcessedURL returns the full URL by substituting any placeholders in the URL.
func (b *Backend) GetProcessedURL(endpointPath string) string {
	return strings.Replace(b.URL, "${path}", endpointPath, -1)
}

// Vhost groups a set of endpoints and specifies any CORS configuration
// that is applied at the vhost level.
type Vhost struct {
	CORS      *CORSConfig `json:"cors,omitempty"`
	Endpoints []Endpoint  `json:"endpoints"`
}

// Config provides a comprehensive view of the API Gateway's configuration,
// encapsulating details about the gateway itself, as well as the vhosts
// and their associated endpoints.
type Config struct {
	APIGateway APIGateway       `json:"apiGateway"`
	Vhosts     map[string]Vhost `json:"vhosts"`
}

// GetConfig reads the API Gateway's configuration from a JSON file and returns it.
// It handles any issues with reading or parsing the configuration file.
func GetConfig() (error, Config) {
	content, err := os.ReadFile("config.json")
	if err != nil {
		L.Fatal("Error reading config.json:", err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		L.Fatal("Error unmarshaling config:", err)
	}
	return err, config
}
