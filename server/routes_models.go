package server

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) registerModelRoutes(r *gin.Engine) {
	// Local model cache management
	r.POST("/api/pull", s.PullHandler)
	r.POST("/api/push", s.PushHandler)
	r.HEAD("/api/tags", s.ListHandler)
	r.GET("/api/tags", s.ListHandler)
	r.POST("/api/show", s.ShowHandler)
	r.DELETE("/api/delete", s.DeleteHandler)
}
