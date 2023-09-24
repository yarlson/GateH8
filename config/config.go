package config

import (
	"encoding/json"
	"fmt"
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

// WebSocketConfig contains configurations specific to WebSocket proxying.
// This includes the backend service URL, as well as the read and write buffer sizes.
type WebSocketConfig struct {
	ReadBufferSize  int      `json:"readBufferSize"`
	WriteBufferSize int      `json:"writeBufferSize"`
	AllowedOrigins  []string `json:"allowedOrigins"`
}

// Endpoint represents a specific route or API endpoint, detailing its
// path, the supported methods, backend service configuration, and any
// CORS policies specific to this endpoint.
type Endpoint struct {
	CORS      *CORSConfig      `json:"cors,omitempty"`
	Path      string           `json:"path"`
	Methods   []string         `json:"methods"`
	Backend   *Backend         `json:"backend"`
	WebSocket *WebSocketConfig `json:"websocket,omitempty"`
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

// Vhost groups a set of endpoints and specifies any CORS and TLS configuration
// that is applied at the vhost level.
type Vhost struct {
	CORS      *CORSConfig `json:"cors,omitempty"`
	Endpoints []Endpoint  `json:"endpoints"`
	TLS       *TLSConfig  `json:"tls,omitempty"`
}

// TLSConfig defines the TLS certificate and key files to be used by the API Gateway.
type TLSConfig struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

// Config provides a comprehensive view of the API Gateway's configuration,
// encapsulating details about the gateway itself, as well as the vhosts
// and their associated endpoints.
type Config struct {
	APIGateway APIGateway       `json:"apiGateway"`
	Vhosts     map[string]Vhost `json:"vhosts"`
	UseTLS     bool
}

// GetConfig reads the API Gateway's configuration from a JSON file and returns it.
// It handles any issues with reading or parsing the configuration file.
func GetConfig() (*Config, error) {
	content, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("error reading config.json: %w", err)
	}

	config := new(Config) // Use new() to get a pointer directly, prevent local-to-heap migration.
	if err = json.Unmarshal(content, config); err != nil {
		return nil, fmt.Errorf("error parsing config.json: %w", err)
	}

	anyVhostWithSSL, allVhostsWithSSL := check(config)

	// If there's any vhost with TLS configured but not all of them have, then it's a config error.
	if anyVhostWithSSL && !allVhostsWithSSL {
		return nil, fmt.Errorf("configuration error: either all vhosts should have TLS configured, or none should")
	}

	config.UseTLS = anyVhostWithSSL
	return config, nil
}

func check(config *Config) (bool, bool) {
	// Check if any vhost has TLS configured
	anyVhostWithSSL := false
	allVhostsWithSSL := true

	for _, vhost := range config.Vhosts {
		if vhost.TLS != nil {
			anyVhostWithSSL = true
		} else {
			allVhostsWithSSL = false
		}
	}
	return anyVhostWithSSL, allVhostsWithSSL
}
