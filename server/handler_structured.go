package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/api"
)

// StructuredHandler handles structured output generation requests.
func (s *Server) StructuredHandler(c *gin.Context) {
	var req api.StructuredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	if req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}

	opts := StructuredOpts{
		Model:      req.Model,
		Prompt:     req.Prompt,
		Schema:     req.Schema,
		MaxRetries: req.MaxRetries,
		Options:    req.Options,
	}

	res, err := s.generateStructured(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := api.StructuredResponse{
		Output:   res.Output,
		Valid:    res.Valid,
		Attempts: res.Attempts,
		Errors:   res.Errors,
	}

	c.JSON(http.StatusOK, resp)
}
