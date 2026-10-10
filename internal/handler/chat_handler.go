package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"linghub/internal/service"
)

// ChatHandler handles chat endpoints.
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// SendMessageRequest holds a message sending request.
type SendMessageRequest struct {
	Message string `json:"message" binding:"required,max=5000"`
}

// SendMessage sends a message to the AI and streams the response via SSE.
// @Summary Send a message and get AI response (SSE)
// @Tags chat
// @Accept json
// @Produce text/event-stream
// @Param id path int true "Conversation ID"
// @Param request body SendMessageRequest true "Send message request"
// @Success 200 {object} string "Server-sent events"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/conversations/:id/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx := c.Request.Context()

	// Stream the chat response
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Callback to send each chunk via SSE
	streamCallback := func(chunk string) error {
		// Send the chunk as a JSON event
		event := map[string]string{"delta": chunk}
		data, _ := json.Marshal(event)
		c.Writer.WriteString("data: ")
		c.Writer.Write(data)
		c.Writer.WriteString("\n\n")
		flusher.Flush()
		return nil
	}

	// Stream the response
	if err := h.chatService.StreamChatResponse(ctx, uint(conversationID), userID.(uint), req.Message, streamCallback); err != nil {
		// If streaming has already started, we can't send a regular error response
		// Instead send it as an SSE event
		c.Writer.WriteString("data: {\"error\": \"")
		c.Writer.WriteString(err.Error())
		c.Writer.WriteString("\"}\n\n")
		flusher.Flush()
		return
	}

	// Send completion marker
	c.Writer.WriteString("data: [DONE]\n\n")
	flusher.Flush()
}
