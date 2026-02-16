package telemetry

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	kafkaMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total number of Kafka messages processed.",
		},
		[]string{"result"},
	)
	kafkaDLQPublishFailuresTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_dlq_publish_failures_total",
			Help: "Total number of DLQ publish failures.",
		},
	)
	storageOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_ops_total",
			Help: "Total number of storage read/write operations.",
		},
		[]string{"store", "op", "result"},
	)
	repositoryUp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repository_up",
			Help: "Repository availability status (1 = up, 0 = down).",
		},
		[]string{"repo"},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		kafkaMessagesTotal,
		kafkaDLQPublishFailuresTotal,
		storageOpsTotal,
		repositoryUp,
	)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			path := r.URL.Path
			if route := mux.CurrentRoute(r); route != nil {
				if tmpl, err := route.GetPathTemplate(); err == nil {
					path = tmpl
				}
			}
			status := strconv.Itoa(rec.status)
			httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			httpRequestDuration.WithLabelValues(r.Method, path, status).Observe(time.Since(start).Seconds())
		}()

		next.ServeHTTP(rec, r)
	})
}

func IncKafkaResult(result string) {
	kafkaMessagesTotal.WithLabelValues(result).Inc()
}

func IncKafkaDLQPublishFailure() {
	kafkaDLQPublishFailuresTotal.Inc()
}

func IncStorageOp(store, op, result string) {
	storageOpsTotal.WithLabelValues(store, op, result).Inc()
}

func SetRepositoryUp(repo string, up bool) {
	value := 0.0
	if up {
		value = 1.0
	}
	repositoryUp.WithLabelValues(repo).Set(value)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
