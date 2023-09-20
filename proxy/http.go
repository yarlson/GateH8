package proxy

import (
	"github.com/yarlson/GateH8/client"
	"github.com/yarlson/GateH8/config"
	"github.com/yarlson/GateH8/logger"
	"io"
	"net/http"
	"strings"
	"time"
)

func CreateHttpProxyHandler(backend *config.Backend, httpClient *http.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := setupRequest(r, backend)
		if err != nil {
			logger.L.Error("Error setting up request:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		proxyClient := client.NewHttpProxyClient(httpClient)
		resp, err := proxyClient.Execute(req, time.Duration(backend.Timeout)*time.Second)
		if err != nil {
			logger.L.Error("Error executing proxy request:", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}

		relayResponse(w, resp)
	}
}

func setupRequest(r *http.Request, backend *config.Backend) (*http.Request, error) {
	processedURL := processURL(backend, r.URL.Path)
	return createRequest(r, processedURL)
}

// relayResponse takes the backend response and relays it back to the original caller.
func relayResponse(w http.ResponseWriter, resp *http.Response) {
	// Business Logic: Relay all headers and the body from the backend response to the original caller
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	body, _ := io.ReadAll(resp.Body)
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func processURL(backend *config.Backend, path string) string {
	if strings.Contains(backend.URL, "${path}") {
		return backend.GetProcessedURL(path)
	}
	return backend.URL
}

func createRequest(r *http.Request, url string) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		return nil, err
	}

	originalUserAgent := r.Header.Get("User-Agent")
	modifiedUserAgent := originalUserAgent + " via GateH8"
	req.Header.Set("User-Agent", modifiedUserAgent)

	return req, nil
}
