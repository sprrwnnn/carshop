package middleware

import (
	"carshop/internal/config"
	"carshop/internal/utils"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

type MetricsMiddleware struct {
	httpRequestsTotal         *prometheus.CounterVec
	httpResponsesTotal        *prometheus.CounterVec
	httpRequestsDuration      *prometheus.HistogramVec
	httpRequestsDurationTotal *prometheus.GaugeVec
	httpRequestsInProgress    *prometheus.GaugeVec
	httpRequestsPanic         prometheus.Counter
}

var (
	pathLabel       = "path"
	methodLabel     = "method"
	statusCodeLabel = "status_code"
)

// TODO: use prometheus.Registerer
//
//nolint:funlen // just a bunch of different metrics
func NewMetricsMiddleware(env config.Env, buildCfg config.BuildInfo) *MetricsMiddleware {
	constLabels := prometheus.Labels{"app_name": buildCfg.Name}

	appInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application info",
		},
		[]string{
			"app_name",
			"version",
			"environment",
			"commit_hash",
			"build_timestamp",
			// TODO: add more info
		},
	)

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests by [Method] [Path].",
			ConstLabels: constLabels,
		},
		[]string{pathLabel, methodLabel},
	)

	responsesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_responses_total",
			Help:        "Total number of HTTP responses by [Method] [Path] [StatusCode].",
			ConstLabels: constLabels,
		},
		[]string{pathLabel, methodLabel, statusCodeLabel},
	)

	requestsDurationTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "http_requests_duration_seconds_total",
			Help:        "Total duration of HTTP requests by [Method] [Path] in seconds.",
			ConstLabels: constLabels,
		},
		[]string{pathLabel, methodLabel},
	)

	requestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Duration of HTTP requests by [Method] [Path] in seconds.",
			ConstLabels: constLabels,
			Buckets:     prometheus.ExponentialBuckets(0.05, 2.0, 10),
		},
		[]string{pathLabel, methodLabel, statusCodeLabel},
	)

	requestsInProgress := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "http_requests_in_progress",
			Help:        "Current amount of HTTP requests being handled",
			ConstLabels: constLabels,
		},
		[]string{pathLabel, methodLabel},
	)

	requestsPanic := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:        "http_requests_panic",
			Help:        "Amount of HTTP requests panicked during handling",
			ConstLabels: constLabels,
		},
	)

	prometheus.MustRegister(appInfo, requestsTotal, responsesTotal, requestsDurationTotal, requestsDuration, requestsInProgress, requestsPanic)

	appInfo = registerMetric(appInfo)
	requestsTotal = registerMetric(requestsTotal)
	responsesTotal = registerMetric(responsesTotal)
	requestsDurationTotal = registerMetric(requestsDurationTotal)
	requestsDuration = registerMetric(requestsDuration)
	requestsInProgress = registerMetric(requestsInProgress)
	requestsPanic = registerMetric(requestsPanic)

	appInfo.WithLabelValues(
		buildCfg.Name,
		buildCfg.Version,
		env,
		buildCfg.Commit,
		buildCfg.Date,
	).Set(1)

	requestsTotal.Reset()
	responsesTotal.Reset()
	requestsDuration.Reset()
	requestsDurationTotal.Reset()
	requestsInProgress.Reset()
	requestsPanic.Write(&io_prometheus_client.Metric{
		Counter: &io_prometheus_client.Counter{
			Value: utils.Ptr(.0),
		},
	})

	return &MetricsMiddleware{
		httpRequestsTotal:         requestsTotal,
		httpResponsesTotal:        responsesTotal,
		httpRequestsDuration:      requestsDuration,
		httpRequestsDurationTotal: requestsDurationTotal,
		httpRequestsInProgress:    requestsInProgress,
		httpRequestsPanic:         requestsPanic,
	}
}

func registerMetric[T prometheus.Collector](collector T) T {
	if err := prometheus.Register(collector); err != nil {
		are := &prometheus.AlreadyRegisteredError{}
		if errors.As(err, are) {
			return are.ExistingCollector.(T)
		}

		panic(err)
	}

	return collector
}

func (m *MetricsMiddleware) Intercept(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		m.httpRequestsTotal.WithLabelValues(path, r.Method).Inc()
		m.httpRequestsInProgress.WithLabelValues(path, r.Method).Inc()

		defer func() {
			m.httpRequestsInProgress.WithLabelValues(path, r.Method).Dec()

			if err := recover(); err != nil {
				m.httpRequestsPanic.Inc()
				m.httpResponsesTotal.WithLabelValues(path, r.Method, "500").Inc()
				panic(err)
			}
		}()

		startTime := time.Now()

		wrappedW := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrappedW, r)

		m.httpResponsesTotal.WithLabelValues(path, r.Method, strconv.Itoa(wrappedW.statusCode)).Inc()

		duration := time.Since(startTime)
		m.httpRequestsDuration.WithLabelValues(path, r.Method, strconv.Itoa(wrappedW.statusCode)).Observe(duration.Seconds())
		m.httpRequestsDurationTotal.WithLabelValues(path, r.Method).Add(duration.Seconds())
	})
}
