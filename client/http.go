package client

import (
	"net/http"
	"time"
)

// HttpProxyClient is responsible for executing the actual HTTP request.
type HttpProxyClient struct {
	client *http.Client
}

// NewHttpProxyClient initializes a new HttpProxyClient with an HTTPClient and a logger.
func NewHttpProxyClient(client *http.Client) *HttpProxyClient {
	return &HttpProxyClient{
		client: client,
	}
}

// Execute sends the HTTP request to the backend and returns the response.
// It uses a client with a timeout.
func (pc *HttpProxyClient) Execute(req *http.Request, timeout time.Duration) (*http.Response, error) {
	// Business Logic: Create a new client with a timeout and execute the request
	clientWithTimeout := &http.Client{Timeout: timeout}
	return clientWithTimeout.Do(req)
}
