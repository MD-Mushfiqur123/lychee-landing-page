package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lychee/lychee/api"
)

// ListConversationsHandler returns a list of all conversations in the store.
func (s *Server) ListConversationsHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	summaries, err := s.memoryStore.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if summaries == nil {
		summaries = []ConversationSummary{}
	}

	c.JSON(http.StatusOK, summaries)
}

// GetConversationHandler loads a specific conversation by ID.
func (s *Server) GetConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	conv, err := s.memoryStore.Load(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// CreateConversationHandler creates a new conversation from scratch.
func (s *Server) CreateConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	var req api.ConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conv := &Conversation{
		ID:        uuid.New().String(),
		Model:     req.Model,
		Messages:  req.Messages,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if conv.Messages == nil {
		conv.Messages = []api.Message{}
	}

	if err := s.memoryStore.Save(conv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, conv)
}

// DeleteConversationHandler deletes a conversation.
func (s *Server) DeleteConversationHandler(c *gin.Context) {
	if s.memoryStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "memory store not initialized"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := s.memoryStore.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
