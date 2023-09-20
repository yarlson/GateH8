package client

import (
	"github.com/gorilla/websocket"
	"github.com/yarlson/GateH8/config"
	"github.com/yarlson/GateH8/logger"
)

// WebSocketProxyClient is a client that handles the proxying of messages between a client
// and a backend WebSocket service. It manages the initial connection with the backend
// and the bidirectional message relay.
type WebSocketProxyClient struct {
	endpoint   config.Endpoint
	clientConn *websocket.Conn
}

// NewWebSocketProxyClient initializes a new WebSocket proxy client. The client takes care of
// establishing a connection with the backend and relaying messages to and from the client.
func NewWebSocketProxyClient(endpoint config.Endpoint, clientConn *websocket.Conn) *WebSocketProxyClient {
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
		logger.L.Error("Failed to establish a WebSocket connection with the backend:", err)
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
			logger.L.Error("Error occurred while reading a message for relay:", err)
			break
		}
		err = dst.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			logger.L.Error("Error occurred while sending a relayed message:", err)
			break
		}
	}
}
