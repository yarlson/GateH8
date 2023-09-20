package proxy

import (
	"github.com/gorilla/websocket"
	"github.com/yarlson/GateH8/client"
	"github.com/yarlson/GateH8/config"
	"github.com/yarlson/GateH8/logger"
	"net/http"
)

// CreateWebSocketProxyHandler creates a handler that handles incoming WebSocket
// connections from clients. This handler is primarily responsible for setting up the initial
// connection but delegates the actual message handling to the WebSocketProxyClient.
func CreateWebSocketProxyHandler(endpoint config.Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set up the WebSocket connection with the proxyClient using predefined parameters.
		// This establishes a full-duplex communication channel between the proxyClient and the proxy server.
		upgrader := getWebSocketUpgrader(endpoint)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.L.Error("Failed to establish a WebSocket connection with the proxyClient:", err)
			return
		}
		logger.L.Info("The connection has been upgraded")
		defer conn.Close()

		// The actual business logic of relaying messages between the proxyClient and a backend
		// WebSocket service is managed by the WebSocketProxyClient.
		proxyClient := client.NewWebSocketProxyClient(endpoint, conn)
		proxyClient.HandleProxy()
	}
}

// getWebSocketUpgrader returns a configured WebSocket upgrader. The upgrader handles the
// specifics of upgrading an HTTP connection to a WebSocket connection based on predefined parameters.
func getWebSocketUpgrader(endpoint config.Endpoint) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  endpoint.WebSocket.ReadBufferSize,
		WriteBufferSize: endpoint.WebSocket.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			// Check the origin of the request to ensure it comes from a trusted source.
			// This prevents unauthorized access and potential security breaches.
			origin := r.Header.Get("Origin")
			for _, allowed := range endpoint.WebSocket.AllowedOrigins {
				if allowed == "*" || origin == allowed {
					return true
				}
			}
			logger.L.Error()
			return false
		},
	}
}
