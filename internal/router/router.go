package router

import (
	"github.com/gin-gonic/gin"
	"linghub/internal/handler"
	"linghub/internal/jwtutil"
	"linghub/internal/middleware"
	"linghub/internal/service"
)

// SetupRouter configures all routes for the application.
func SetupRouter(
	authService *service.AuthService,
	chatService *service.ChatService,
	tm *jwtutil.TokenManager,
) *gin.Engine {
	router := gin.Default()

	// Create handlers
	authHandler := handler.NewAuthHandler(authService)
	conversationHandler := handler.NewConversationHandler(chatService)
	chatHandler := handler.NewChatHandler(chatService)

	// Public routes (no authentication required)
	public := router.Group("/api")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/api/auth/login", authHandler.Login)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api").Use(middleware.Auth(tm))
	{
		// Conversation routes
		protected.POST("/conversations", conversationHandler.CreateConversation)
		protected.GET("/conversations", conversationHandler.GetConversations)
		protected.GET("/conversations/:id/messages", conversationHandler.GetMessages)

		// Chat routes
		protected.POST("/conversations/:id/messages", chatHandler.SendMessage)
	}

	return router
}
