package internal

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// CreateProxyHandler generates a handler to proxy requests to a specified backend.
// It dynamically adjusts the destination URL if placeholders (like "${path}") are detected.
// Once the backend is called, the response headers and body are relayed back to the original caller.
// Additionally, a modification is done to the 'User-Agent' header of the request
// to include an identifier of the proxy, "via GateH8".
func CreateProxyHandler(backend *Backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Initializing an HTTP client with a timeout specified by the backend config.
		client := &http.Client{
			Timeout: time.Duration(backend.Timeout) * time.Millisecond,
		}

		// Adjust the backend URL if there's a placeholder for the path.
		processedURL := backend.URL
		if strings.Contains(backend.URL, "${path}") {
			processedURL = backend.GetProcessedURL(r.URL.Path)
		}
		req, err := http.NewRequest(r.Method, processedURL, r.Body)
		if err != nil {
			L.Error("Error creating request to backend service:", err)
			http.Error(w, "Error creating request to backend service", http.StatusInternalServerError)
			return
		}

		// Modify the 'User-Agent' to include the proxy's identifier.
		originalUserAgent := r.Header.Get("User-Agent")
		modifiedUserAgent := originalUserAgent + " via GateH8"
		req.Header.Set("User-Agent", modifiedUserAgent)

		// Send the request to the backend and handle any errors.
		resp, err := client.Do(req)
		if err != nil {
			L.Error("Error calling backend service:", err)
			http.Error(w, "Error calling backend service", http.StatusBadGateway)
			return
		}
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

		// Relay all headers from the backend response to the original caller.
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Relay the backend response's body to the original caller.
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)
	}
}
