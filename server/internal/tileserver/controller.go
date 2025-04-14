package tileserver

import (
	"net/http"
	"server/internal/database"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ITileServerController interface {
	GetTile(c *gin.Context)
	GetFeatures(c *gin.Context)
	SearchLocation(c *gin.Context)
	GetMapInfo(c *gin.Context)
}

type tileServer struct {
	logger     *zap.Logger
	repository Repository
}

func NewTileServerController(db database.Service ,logger *zap.Logger) ITileServerController {
	repo := NewTileRepository(db,logger)
	return &tileServer{
		logger:     logger,
		repository: repo,
	}
}

// -------------------- GET TILE --------------------

func (t *tileServer) GetTile(c *gin.Context) {
	zStr := c.Param("z")
	xStr := c.Param("x")
	yStr := strings.TrimSuffix(c.Param("y.mvt"), ".mvt")

	z, err := strconv.Atoi(zStr)
	x, errX := strconv.Atoi(xStr)
	y, errY := strconv.Atoi(yStr)

	if err != nil || errX != nil || errY != nil {
		t.logger.Error("Invalid tile coordinates", zap.String("z", zStr), zap.String("x", xStr), zap.String("y", yStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tile coordinates"})
		return
	}

	tileData, err := t.repository.GetTile(z, x, y)
	if err != nil {
		t.logger.Error("Failed to get tile data", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Type", "application/x-protobuf")
	c.Data(http.StatusOK, "application/x-protobuf", tileData)
}

// -------------------- GET FEATURES --------------------

func (t *tileServer) GetFeatures(c *gin.Context) {
	bbox := c.Query("bbox")
	zoomStr := c.Query("zoom")

	if bbox == "" || zoomStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing bbox or zoom parameters"})
		return
	}

	bounds := strings.Split(bbox, ",")
	if len(bounds) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bbox format"})
		return
	}

	zoom, err := strconv.Atoi(zoomStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid zoom level"})
		return
	}

	features, err := t.repository.GetFeatures(bounds, zoom)
	if err != nil {
		t.logger.Error("Failed to fetch features", zap.Error(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Cache-Control", "public, max-age=600")
	c.JSON(http.StatusOK, features)
}

// -------------------- SEARCH LOCATION --------------------

func (t *tileServer) SearchLocation(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No search query provided"})
		return
	}

	results, err := t.repository.SearchLocation(query)
	if err != nil {
		t.logger.Error("Search query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// -------------------- GET MAP INFO --------------------

func (t *tileServer) GetMapInfo(c *gin.Context) {
	info, err := t.repository.GetMapInfo()
	if err != nil {
		t.logger.Error("Failed to retrieve map info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get map info"})
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.JSON(http.StatusOK, info)
}
