package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/DostonAkhmedov/task-manager/util/metrics"
	"github.com/DostonAkhmedov/task-manager/internal/service"
	"github.com/DostonAkhmedov/task-manager/internal/transport/handlers"
)

// commonMiddlewares returns a slice of global middleware functions
func commonMiddlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	}
}

// SetupRouter configures and returns the HTTP router
func (a *App) SetupRouter() chi.Router {
	router := chi.NewRouter()

	// Apply global middleware
	router.Use(commonMiddlewares()...)
	// Apply rate limiting globally
	router.Use(a.rateLimiter.Middleware())
	// Apply custom metrics middleware
	router.Use(metrics.HTTPMetricsMiddleware(a.metrics))

	// Metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	// Initialize services with configuration
	teamService := service.NewTeamService(a.teamRepo, a.userRepo)

	cacheTTL := time.Duration(a.cfg.Redis.CacheTTL) * time.Minute
	taskService := service.NewTaskService(a.taskRepo, a.teamRepo, a.redisClient, cacheTTL)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(a.userRepo, a.jwtManager)
	teamHandler := handlers.NewTeamHandler(teamService)
	taskHandler := handlers.NewTaskHandler(taskService)
	queryHandler := handlers.NewComplexQueryHandler(a.complexQueryRepo)

	// Public routes (no authentication required)
	router.Post("/api/v1/register", authHandler.Register)
	router.Post("/api/v1/login", authHandler.Login)

	// Public query endpoints (analytics, no auth required for now)
	router.Get("/api/v1/queries/team-statistics", queryHandler.GetTeamStatistics)
	router.Get("/api/v1/queries/top-task-creators", queryHandler.GetTopTaskCreators)
	router.Get("/api/v1/queries/invalid-assignments", queryHandler.GetInvalidAssignments)

	// Protected routes (authentication required)
	router.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware to all protected routes
		r.Use(authMiddleware(a.jwtManager))

		// Team endpoints
		r.Route("/teams", func(r chi.Router) {
			r.Post("/", teamHandler.CreateTeam)
			r.Get("/", teamHandler.GetTeams)
			r.Get("/{id}", teamHandler.GetTeam)
			r.Get("/{id}/members", teamHandler.GetTeamMembers)
			r.Post("/{id}/invite", teamHandler.InviteTeamMember)
		})

		// Task endpoints
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", taskHandler.CreateTask)
			r.Get("/", taskHandler.GetTasks)
			r.Get("/{id}", taskHandler.GetTask)
			r.Put("/{id}", taskHandler.UpdateTask)
			r.Get("/{id}/history", taskHandler.GetTaskHistory)
		})
	})

	return router
}
