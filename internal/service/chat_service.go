package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"gorm.io/gorm"
	"linghub/internal/config"
	"linghub/internal/model"
)

// ChatService handles AI chat conversations using Eino.
type ChatService struct {
	db     *gorm.DB
	runner *adk.Runner
	cfg    *config.Config
}

// NewChatService creates a new chat service.
func NewChatService(db *gorm.DB, cfg *config.Config) (*ChatService, error) {
	ctx := context.Background()

	// Initialize OpenAI ChatModel
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.OpenAIAPIKey,
		BaseURL: cfg.OpenAIBaseURL,
		Model:   cfg.OpenAIModel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI chat model: %w", err)
	}

	// Create ChatModelAgent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Model:       chatModel,
		Instruction: cfg.SystemPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model agent: %w", err)
	}

	// Create Runner with streaming enabled
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
	})

	return &ChatService{
		db:     db,
		runner: runner,
		cfg:    cfg,
	}, nil
}

// CreateConversation creates a new conversation for a user.
func (s *ChatService) CreateConversation(userID uint, title string) (*model.Conversation, error) {
	conversation := &model.Conversation{
		UserID: userID,
		Title:  title,
	}

	if err := s.db.Create(conversation).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// GetUserConversations retrieves all conversations for a user.
func (s *ChatService) GetUserConversations(userID uint) ([]model.Conversation, error) {
	var conversations []model.Conversation
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch conversations: %w", err)
	}

	return conversations, nil
}

// GetConversationMessages retrieves all messages in a conversation.
func (s *ChatService) GetConversationMessages(conversationID uint) ([]model.Message, error) {
	var messages []model.Message
	if err := s.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return messages, nil
}

// ChatStreamRequest holds a chat request.
type ChatStreamRequest struct {
	ConversationID uint   `json:"conversation_id" binding:"required"`
	Message        string `json:"message" binding:"required,max=5000"`
}

// ChatMessage represents a message to send to the model.
type ChatMessage struct {
	Role    string
	Content string
}

// StreamChatResponse streams a chat response using SSE.
// It reads from the chat model and yields text chunks via the callback.
func (s *ChatService) StreamChatResponse(ctx context.Context, conversationID, userID uint, userMessage string, streamCallback func(string) error) error {
	// Verify conversation ownership
	var conversation model.Conversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("conversation not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Get conversation history
	dbMessages, err := s.GetConversationMessages(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Convert to schema.Message for Eino
	schemaMessages := make([]*schema.Message, 0, len(dbMessages)+1)
	for _, msg := range dbMessages {
		var role schema.RoleType
		if msg.Role == "user" {
			role = schema.User
		} else {
			role = schema.Assistant
		}
		schemaMessages = append(schemaMessages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add the new user message
	schemaMessages = append(schemaMessages, &schema.Message{
		Role:    schema.User,
		Content: userMessage,
	})

	// Run the agent
	iter := s.runner.Run(ctx, schemaMessages)
	defer iter.Next() // Ensure iterator is properly closed

	// Collect the full AI response
	var aiResponseBuilder strings.Builder

	// Process events from the iterator
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			return fmt.Errorf("agent error: %w", event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msgVariant := event.Output.MessageOutput

			if msgVariant.IsStreaming && msgVariant.MessageStream != nil {
				// Stream mode: read chunks and send via callback
				for {
					chunk, err := msgVariant.MessageStream.Recv()
					if err != nil && err != io.EOF {
						return fmt.Errorf("stream error: %w", err)
					}
					if err == io.EOF {
						break
					}

					if chunk != nil && chunk.Content != "" {
						aiResponseBuilder.WriteString(chunk.Content)
						if err := streamCallback(chunk.Content); err != nil {
							return fmt.Errorf("callback error: %w", err)
						}
					}
				}
			} else if msgVariant.Message != nil && msgVariant.Message.Content != "" {
				// Non-stream mode: use the complete message
				content := msgVariant.Message.Content
				aiResponseBuilder.WriteString(content)
				if err := streamCallback(content); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
		}
	}

	// Save the exchange to the database
	userMsg := &model.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        userMessage,
	}
	if err := s.db.Create(userMsg).Error; err != nil {
		return fmt.Errorf("failed to save user message: %w", err)
	}

	assistantMsg := &model.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        aiResponseBuilder.String(),
	}
	if err := s.db.Create(assistantMsg).Error; err != nil {
		return fmt.Errorf("failed to save assistant message: %w", err)
	}

	return nil
}
