package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
	Close() error
	GetDB() *sql.DB
}

type service struct {
	db     *sql.DB
	logger *zap.Logger
}

var (
	dbInstance *service

	dbName     = os.Getenv("DB_DATABASE")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbUser     = os.Getenv("DB_USERNAME")
	dbPort     = os.Getenv("DB_PORT")
	dbHost     = os.Getenv("DB_HOST")
)

// New returns a singleton DB service instance with connection pooling.
func New(logger *zap.Logger) Service {
	if dbInstance != nil {
		return dbInstance
	}

	portInt, err := strconv.Atoi(dbPort)
	if err != nil {
		logger.Fatal("unable to convert DB_PORT to integer", zap.Error(err))
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, portInt, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err = db.Ping(); err != nil {
		logger.Fatal("database ping failed", zap.Error(err))
	}

	dbInstance = &service{
		db:     db,
		logger: logger,
	}

	logger.Info("Connected to database successfully")
	return dbInstance
}

func (s *service) GetDB() *sql.DB {
	return s.db
}

func (s *service) Close() error {
	s.logger.Info("Disconnected from database")
	return s.db.Close()
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "Database is healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	return stats
}
