package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AliasRequest struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// SetAliasHandler handles creating or updating a model alias.
func (s *Server) SetAliasHandler(c *gin.Context) {
	if s.modelAliases == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model aliases not initialized"})
		return
	}

	var req AliasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.modelAliases.Set(req.Name, req.Target); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "name": req.Name, "target": req.Target})
}

// ListAliasesHandler handles listing all registered model aliases.
func (s *Server) ListAliasesHandler(c *gin.Context) {
	if s.modelAliases == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model aliases not initialized"})
		return
	}

	aliases := s.modelAliases.List()
	if aliases == nil {
		aliases = make(map[string]string)
	}
	c.JSON(http.StatusOK, aliases)
}

// DeleteAliasHandler handles deleting a model alias.
func (s *Server) DeleteAliasHandler(c *gin.Context) {
	if s.modelAliases == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model aliases not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := s.modelAliases.Delete(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
