package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LoginAttemptTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_attempt_total",
		Help: "Total number of login attempts.",
	})

	LoginSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_success_total",
		Help: "Total number of successful logins.",
	})

	LoginFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_failure_total",
		Help: "Total number of failed login attempts.",
	})

	TokenRefreshTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "token_refresh_total",
		Help: "Total number of token refresh operations.",
	})

	AuthRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_request_total",
			Help: "Total number of authenticated requests, labelled by method and path.",
		},
		[]string{"method", "path", "status"},
	)

	AuthRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_request_duration_seconds",
			Help:    "Duration of authenticated HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)
