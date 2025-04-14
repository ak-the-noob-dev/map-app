package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()
	r.Use(gin.Recovery())

	r.Use(s.corsMiddleware())

	// Static file serving - serve compressed files when available
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Serve index.html on root request
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	r.GET("/op", func(c *gin.Context) {
		c.File("./static/openstreet.html")
	})

	r.GET("/health", s.healthHandler)

	s.registerTileServerRoutes(r)

	return r
}

func (s *Server) serveIndex(c *gin.Context) {
	// Get the absolute path to index.html
	indexPath := filepath.Join("html", "index.html")
	
	// Check if the file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "index.html not found"})
		return
	}

	// Serve the file
	c.File(indexPath)
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		c.Header("Access-Control-Allow-Credentials", "false")

		// Handle preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		c.Next()
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal health check response"})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", resp)
}

func (s *Server) registerTileServerRoutes(r *gin.Engine) {
	// API Routes
	r.GET("/api/map-info", s.tileserver.GetMapInfo)
	r.GET("/api/features", s.tileserver.GetFeatures)
	r.GET("/api/search", s.tileserver.SearchLocation)
	
	// Tile endpoints
	r.GET("/tiles/:z/:x/:y.mvt", s.tileserver.GetTile)

}
