package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"linghub/internal/service"
)

// ConversationHandler handles conversation endpoints.
type ConversationHandler struct {
	chatService *service.ChatService
}

// NewConversationHandler creates a new conversation handler.
func NewConversationHandler(chatService *service.ChatService) *ConversationHandler {
	return &ConversationHandler{chatService: chatService}
}

// CreateConversationRequest holds a conversation creation request.
type CreateConversationRequest struct {
	Title string `json:"title" binding:"required,max=255"`
}

// CreateConversation creates a new conversation.
// @Summary Create a new conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Param request body CreateConversationRequest true "Create conversation request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/conversations [post]
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation, err := h.chatService.CreateConversation(userID.(uint), req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    conversation.ID,
		"title": conversation.Title,
	})
}

// GetConversations retrieves all conversations for the user.
// @Summary Get user conversations
// @Tags conversations
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/conversations [get]
func (h *ConversationHandler) GetConversations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversations, err := h.chatService.GetUserConversations(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// GetMessages retrieves messages in a conversation.
// @Summary Get conversation messages
// @Tags conversations
// @Produce json
// @Param id path int true "Conversation ID"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/conversations/:id/messages [get]
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation ID"})
		return
	}

	messages, err := h.chatService.GetConversationMessages(uint(conversationID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}
