package server

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lychee/lychee/server/webui"
)

// RegisterDashboardRoutes registers the dashboard UI endpoints under /dashboard/*.
func (s *Server) registerDashboardRoutes(r *gin.Engine) {
	subFS, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(subFS))

	r.GET("/dashboard/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		filepath = strings.TrimPrefix(filepath, "/")

		// Check if the requested file exists in the embedded FS.
		// If not, or if it's a directory, fallback to index.html for client-side routing.
		if filepath == "" || filepath == "/" {
			c.FileFromFS("index.html", http.FS(subFS))
			return
		}

		f, err := subFS.Open(filepath)
		if err != nil {
			// File does not exist, serve index.html for client-side routing
			c.FileFromFS("index.html", http.FS(subFS))
			return
		}
		f.Close()

		// Serve the file
		c.Request.URL.Path = filepath
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
