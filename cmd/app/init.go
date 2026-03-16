package app

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	"github.com/DostonAkhmedov/task-manager/internal/cache"
	"github.com/DostonAkhmedov/task-manager/internal/database"
	util "github.com/DostonAkhmedov/task-manager/util/auth"
)

// NewDatabase creates and initializes a MySQL database connection
// It configures the connection pool and verifies connectivity
func NewDatabase(cfg *DatabaseConfig) (*database.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for optimal performance
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MinConns)

	// Verify the connection is working
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &database.DB{DB: db}, nil
}

// NewRedis creates and initializes a Redis client connection
func NewRedis(cfg *RedisConfig) (*cache.Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &cache.Redis{Client: client}, nil
}

// NewJWTManager creates a JWT manager configured with the specified secret and expiration
func NewJWTManager(cfg *JWTConfig) *util.JWTManager {
	return util.NewJWTManagerWithSecret(cfg.Secret, cfg.Expiration)
}
