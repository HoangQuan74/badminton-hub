package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	websocketValidation "github.com/kainguyen/badminton-hub/be/gateways/socket/transport/http/validation"
	socketV1 "github.com/kainguyen/badminton-hub/contracts/gen/socket/v1"
)

type WebSocket struct {
	client   socketV1.SocketServiceClient
	upgrader websocket.Upgrader
}

func NewWebSocket(client socketV1.SocketServiceClient) *WebSocket {
	return &WebSocket{client: client, upgrader: websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}}
}

func (h *WebSocket) Handle(c *gin.Context) {
	accessToken, valid := websocketValidation.AccessToken(c.GetHeader("Authorization"))
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required."}})
		return
	}
	connection, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer connection.Close()
	connection.SetReadLimit(websocketValidation.MaxMessageSize)
	stream, err := h.client.Connect(c.Request.Context())
	if err != nil {
		return
	}
	if err := stream.Send(&socketV1.ClientMessage{AccessToken: accessToken}); err != nil {
		return
	}
	go h.writeMessages(connection, stream)
	h.readMessages(connection, stream)
}

func (h *WebSocket) writeMessages(connection *websocket.Conn, stream socketV1.SocketService_ConnectClient) {
	for {
		message, err := stream.Recv()
		if err != nil {
			return
		}
		messageType := websocket.TextMessage
		if message.GetBinary() {
			messageType = websocket.BinaryMessage
		}
		if err := connection.WriteMessage(messageType, message.GetPayload()); err != nil {
			return
		}
	}
}

func (h *WebSocket) readMessages(connection *websocket.Conn, stream socketV1.SocketService_ConnectClient) {
	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			return
		}
		if err := stream.Send(&socketV1.ClientMessage{Payload: payload, Binary: messageType == websocket.BinaryMessage}); err != nil {
			return
		}
	}
}

func RegisterWebSocket(engine *gin.Engine, client socketV1.SocketServiceClient, metricsHandler http.Handler) {
	// Build the WebSocket handler with the owning Socket gRPC client.
	handler := NewWebSocket(client)

	// Register public WebSocket and system routes.
	engine.GET("/ws", handler.Handle)
	engine.GET("/ws/*path", handler.Handle)
	engine.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	engine.GET("/ready", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	engine.GET("/metrics", gin.WrapH(metricsHandler))
}
