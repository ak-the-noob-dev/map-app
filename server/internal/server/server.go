package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"server/infra/logger"
	"server/internal/database"
	"server/internal/tileserver"
)

type Server struct {
	port int

	db     database.Service

	tileserver tileserver.ITileServerController
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	logger := logger.NewLogger()
	db := database.New(logger)
	tileserver := tileserver.NewTileServerController(db,logger)

	
	NewServer := &Server{
		port: port,

		db: db,

		tileserver: tileserver,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
