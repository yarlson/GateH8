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
func GetConfig() (error, *Config) {
	content, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("error reading config.json: %w", err), nil
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("error parsing config.json: %w", err), nil
	}

	// Check the TLS configurations for all vhosts
	vhostsWithSSL := 0
	for _, vhost := range config.Vhosts {
		if vhost.TLS != nil {
			vhostsWithSSL++
		}
	}

	// If there are any vhosts with TLS configured, then all vhosts should have TLS configured.
	if vhostsWithSSL > 0 && vhostsWithSSL != len(config.Vhosts) {
		return fmt.Errorf("configuration error: either all vhosts should have TLS configured, or none should"), nil
	}

	// If there are vhosts with TLS configured, then the API Gateway should use TLS.
	if vhostsWithSSL > 0 {
		config.UseTLS = true
	}

	return err, &config
}
