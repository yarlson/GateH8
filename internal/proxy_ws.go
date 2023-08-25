package internal

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// CreateWebSocketProxyHandler creates a handler that handles incoming WebSocket
// connections from clients. This handler is primarily responsible for setting up the initial
// connection but delegates the actual message handling to the WebSocketProxyClient.
func CreateWebSocketProxyHandler(endpoint Endpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set up the WebSocket connection with the client using predefined parameters.
		// This establishes a full-duplex communication channel between the client and the proxy server.
		upgrader := getWebSocketUpgrader(endpoint)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			L.Error("Failed to establish a WebSocket connection with the client:", err)
			return
		}
		defer conn.Close()

		// The actual business logic of relaying messages between the client and a backend
		// WebSocket service is managed by the WebSocketProxyClient.
		client := NewWebSocketProxyClient(endpoint, conn)
		client.HandleProxy()
	}
}

// WebSocketProxyClient is a client that handles the proxying of messages between a client
// and a backend WebSocket service. It manages the initial connection with the backend
// and the bidirectional message relay.
type WebSocketProxyClient struct {
	endpoint   Endpoint
	clientConn *websocket.Conn
}

// NewWebSocketProxyClient initializes a new WebSocket proxy client. The client takes care of
// establishing a connection with the backend and relaying messages to and from the client.
func NewWebSocketProxyClient(endpoint Endpoint, clientConn *websocket.Conn) *WebSocketProxyClient {
	return &WebSocketProxyClient{
		endpoint:   endpoint,
		clientConn: clientConn,
	}
}

// HandleProxy establishes a connection with the backend WebSocket service and initiates
// the bidirectional message relay. It manages two communication channels: one from
// the client to the backend and another from the backend to the client.
func (c *WebSocketProxyClient) HandleProxy() {
	backendConn, _, err := websocket.DefaultDialer.Dial(c.endpoint.Backend.URL, nil)
	if err != nil {
		L.Error("Failed to establish a WebSocket connection with the backend:", err)
		return
	}
	defer backendConn.Close()

	// Set up two communication channels for bidirectional message relay.
	done := make(chan struct{})
	defer close(done)

	// One channel listens to messages from the client and sends them to the backend.
	go c.relayMessages(c.clientConn, backendConn)

	// The other listens to messages from the backend and sends them to the client.
	go c.relayMessages(backendConn, c.clientConn)

	// Wait until both communication channels complete their operations.
	<-done
}

// relayMessages handles the relay of messages between a source and a destination WebSocket.
// It continuously listens for incoming messages from the source and forwards them to the destination.
func (c *WebSocketProxyClient) relayMessages(src, dst *websocket.Conn) {
	for {
		_, message, err := src.ReadMessage()
		if err != nil {
			L.Error("Error occurred while reading a message for relay:", err)
			break
		}
		err = dst.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			L.Error("Error occurred while sending a relayed message:", err)
			break
		}
	}
}

// getWebSocketUpgrader returns a configured WebSocket upgrader. The upgrader handles the
// specifics of upgrading an HTTP connection to a WebSocket connection based on predefined parameters.
func getWebSocketUpgrader(endpoint Endpoint) websocket.Upgrader {
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
			return false
		},
	}
}
