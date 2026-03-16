package app

import (
	"fmt"
	"log"

	"github.com/DostonAkhmedov/task-manager/internal/cache"
	"github.com/DostonAkhmedov/task-manager/internal/database"
	"github.com/DostonAkhmedov/task-manager/util/metrics"
	"github.com/DostonAkhmedov/task-manager/util/ratelimit"
	"github.com/DostonAkhmedov/task-manager/internal/repository"
	util "github.com/DostonAkhmedov/task-manager/util/auth"
)

// App represents the main application with all its dependencies
type App struct {
	cfg               *Config
	db                *database.DB
	redisClient       *cache.Redis
	userRepo          *repository.UserRepository
	teamRepo          *repository.TeamRepository
	taskRepo          *repository.TaskRepository
	jwtManager        *util.JWTManager
	rateLimiter       *ratelimit.RateLimiter
	complexQueryRepo  *repository.ComplexQueryRepository
	metrics           *metrics.Metrics
}

// NewApp creates and initializes a new application
func NewApp() (*App, error) {
	log.Println("Initializing configuration...")
	cfg, err := Load("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log.Println("Initializing database...")
	db, err := NewDatabase(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	log.Println("Initializing Redis cache...")
	redisClient, err := NewRedis(&cfg.Redis)
	if err != nil {
		db.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	teamRepo := repository.NewTeamRepository(db.DB)
	taskRepo := repository.NewTaskRepository(db.DB)
	complexQueryRepo := repository.NewComplexQueryRepository(db.DB)

	// Initialize JWT manager
	jwtManager := NewJWTManager(&cfg.JWT)

	// Initialize rate limiter
	log.Println("Initializing rate limiter...")
	rateLimiter := ratelimit.NewRateLimiter(redisClient.Client, cfg.Server.RateLimit)

	// Initialize metrics
	log.Println("Initializing Prometheus metrics...")
	appMetrics := metrics.NewMetrics()

	return &App{
		cfg:              cfg,
		db:               db,
		redisClient:      redisClient,
		userRepo:         userRepo,
		teamRepo:         teamRepo,
		taskRepo:         taskRepo,
		jwtManager:       jwtManager,
		rateLimiter:      rateLimiter,
		complexQueryRepo: complexQueryRepo,
		metrics:          appMetrics,
	}, nil
}

// Config returns the application configuration
func (a *App) Config() *Config {
	return a.cfg
}

// Close closes all resources and connections
func (a *App) Close() {
	if a.db != nil {
		a.db.Close() //nolint:errcheck
	}
	if a.redisClient != nil {
		a.redisClient.Close() //nolint:errcheck
	}
}
