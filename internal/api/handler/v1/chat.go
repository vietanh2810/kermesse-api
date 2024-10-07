package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"net/http"
	"strconv"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this for production!
	},
}

type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	userRole string
}

type ChatHandler struct {
	svc          KermesseService
	uSvc         UserService
	clients      map[uint]*Client
	clientsMutex sync.RWMutex
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
}

func NewChatHandler(svc KermesseService, uSvc UserService) *ChatHandler {
	return &ChatHandler{
		svc:        svc,
		uSvc:       uSvc,
		clients:    make(map[uint]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *ChatHandler) Run() {
	for {
		select {
		case client := <-h.register:
			h.clientsMutex.Lock()
			h.clients[client.userID] = client
			h.clientsMutex.Unlock()
		case client := <-h.unregister:
			h.clientsMutex.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.clientsMutex.Unlock()
		case message := <-h.broadcast:
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.userID)
				}
			}
		}
	}
}

// HandleWebSocket godoc
// @Summary Establish WebSocket connection for chat
// @Description Establishes a WebSocket connection for real-time chat between organizers and stand holders
// @Tags kermesses,chat
// @Produce json
// @Param kermesseID path int true "Kermesse ID"
// @Param standID path int true "Stand ID"
// @Success 101 {string} string "Switching Protocols to WebSocket"
// @Failure 400 {object} response.Err
// @Failure 401 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/stands/{standID}/chat [get]
// @Security BearerAuth
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	user, respErr := getUserFromContext(c, h.uSvc)
	if respErr != nil {
		conn.Close()
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   user.ID,
		userRole: user.Role,
	}
	h.register <- client

	go client.writePump()
	go client.readPump(h)
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(h *ChatHandler) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}

		var chatMessage domain.ChatMessage
		if err := json.Unmarshal(message, &chatMessage); err != nil {
			fmt.Println(err)
			continue
		}

		chatMessage.SenderID = c.userID

		// Check if the sender is authorized to send the message
		if c.userRole == "organizer" || (c.userRole == "stand_holder" && chatMessage.StandID != 0) {
			savedMessage, err := h.svc.SaveChatMessage(chatMessage)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Send the message to the appropriate recipient
			h.clientsMutex.RLock()
			if recipient, ok := h.clients[chatMessage.ReceiverID]; ok {
				recipient.send <- message
			}
			h.clientsMutex.RUnlock()

			// Send a confirmation back to the sender
			confirmationMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "confirmation",
				"message": "Message sent successfully",
				"id":      savedMessage.ID,
			})
			c.send <- confirmationMsg
		} else {
			errorMsg, _ := json.Marshal(map[string]interface{}{
				"type":    "error",
				"message": "You are not authorized to send this message",
			})
			c.send <- errorMsg
		}
	}
}

// HandleGetChatMessages godoc
// @Summary Get chat messages
// @Description Retrieves chat messages for a specific kermesse and stand
// @Tags kermesses,chat
// @Produce json
// @Param kermesseID path int true "Kermesse ID"
// @Param standID path int true "Stand ID"
// @Param limit query int false "Number of messages to retrieve (default 50)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} domain.ChatMessage
// @Failure 400 {object} response.Err
// @Failure 401 {object} response.Err
// @Failure 403 {object} response.Err
// @Failure 500 {object} response.Err
// @Router /kermesses/{kermesseID}/stands/{standID}/messages [get]
// @Security BearerAuth
func (h *ChatHandler) HandleGetChatMessages(c *gin.Context) {
	kermesseID, _ := strconv.ParseUint(c.Param("kermesseID"), 10, 32)
	standID, _ := strconv.ParseUint(c.Param("standID"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.svc.GetChatMessages(uint(kermesseID), uint(standID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
