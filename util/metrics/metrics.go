package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all custom Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   prometheus.CounterVec
	HTTPRequestDuration prometheus.HistogramVec
	HTTPErrorsTotal     prometheus.CounterVec

	// Task metrics
	TasksCreatedTotal   prometheus.Counter
	TasksUpdatedTotal   prometheus.Counter
	TaskDeletedTotal    prometheus.Counter

	// Team metrics
	TeamsCreatedTotal   prometheus.Counter
	TeamMembersAdded    prometheus.Counter

	// Cache metrics
	CacheHitsTotal      prometheus.Counter
	CacheMissesTotal    prometheus.Counter

	// Database metrics
	DBQueryDuration     prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		HTTPErrorsTotal: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total HTTP errors",
			},
			[]string{"method", "path", "status"},
		),

		TasksCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_created_total",
				Help: "Total tasks created",
			},
		),

		TasksUpdatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_updated_total",
				Help: "Total tasks updated",
			},
		),

		TaskDeletedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "tasks_deleted_total",
				Help: "Total tasks deleted",
			},
		),

		TeamsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "teams_created_total",
				Help: "Total teams created",
			},
		),

		TeamMembersAdded: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "team_members_added_total",
				Help: "Total team members added",
			},
		),

		CacheHitsTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total cache hits",
			},
		),

		CacheMissesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total cache misses",
			},
		),

		DBQueryDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"query_type", "table"},
		),
	}
}
