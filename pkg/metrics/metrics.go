package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrderTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "order_total",
		Help: "Total number of orders created",
	})
	OrderFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "order_failed_total",
		Help: "Total number of failed orders",
	})
	AssetGrantTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "asset_grant_total",
		Help: "Total number of asset grants",
	})
	AssetGrantFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "asset_grant_failed_total",
		Help: "Total number of failed asset grants",
	})
	OutboxPendingTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_pending_total",
		Help: "Current number of pending outbox messages",
	})
	ReconcileRetryTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reconcile_retry_total",
		Help: "Total number of reconcile retries",
	})
	ReconcileFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reconcile_failed_total",
		Help: "Total number of reconcile failures",
	})
	APIDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "api_request_duration_seconds",
		Help:    "API request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
)
