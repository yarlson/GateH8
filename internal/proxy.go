package internal

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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

// CreateWebSocketProxyHandler generates a handler to proxy WebSocket requests to a specified backend.
func CreateWebSocketProxyHandler(endpoint Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  endpoint.WebSocket.ReadBufferSize,
			WriteBufferSize: endpoint.WebSocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				for _, allowed := range endpoint.WebSocket.AllowedOrigins {
					if allowed == "*" {
						return true
					}
					if origin == allowed {
						return true
					}
				}
				return false
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			L.Error("Failed to upgrade the connection:", err)
			return
		}
		defer func(conn *websocket.Conn) { _ = conn.Close() }(conn)

		// Create a connection to the backend WebSocket server.
		backendConn, _, err := websocket.DefaultDialer.Dial(endpoint.Backend.URL, nil)
		if err != nil {
			L.Error("Failed to connect to the backend WebSocket server:", err)
			return
		}
		defer func(backendConn *websocket.Conn) { _ = backendConn.Close() }(backendConn)

		// Channel to signal the main goroutine to exit after both read and write goroutines are done.
		done := make(chan struct{})

		defer close(done)

		// Start goroutine to read from client and write to backend.
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					L.Error("Error reading from WebSocket:", err)
					break
				}
				err = backendConn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					L.Error("Error writing to WebSocket:", err)
					break
				}
			}
		}()

		// Start goroutine to read from backend and write to client.
		go func() {
			for {
				_, message, err := backendConn.ReadMessage()
				if err != nil {
					L.Error("Error reading from WebSocket:", err)
					break
				}
				err = conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					L.Error("Error writing to WebSocket:", err)
					break
				}
			}
		}()

		// Wait for both read and write goroutines to finish.
		<-done
	}
}
