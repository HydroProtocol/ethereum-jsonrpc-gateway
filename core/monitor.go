package core

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var counter *prometheus.CounterVec
var gauge *prometheus.GaugeVec
var histogram *prometheus.HistogramVec

func init() {
	histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "histogram",
		Help: "histogram",
	}, []string{"key"})

	counter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "counter",
		Help: "counter",
	}, []string{"key"})

	gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gauge",
		Help: "gauge",
	}, []string{"key"})

	prometheus.MustRegister(counter)
	prometheus.MustRegister(gauge)
	prometheus.MustRegister(histogram)
}

func Time(key string, value float64) {
	histogram.WithLabelValues(key).Observe(value)
}

func Count(key string) {
	counter.WithLabelValues(key).Inc()
}

func Value(key string, value float64) {
	gauge.WithLabelValues(key).Set(value)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func StartMonitorHttpServer(ctx context.Context) {
	addr := "0.0.0.0:9090"

	hs := &http.Server{
		Addr:    addr,
		Handler: promhttp.Handler(),
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := hs.Shutdown(shutdownCtx); err != nil {
			logrus.Fatalf("Could not gracefully shutdown metric server: %v\n", err)
		}
	}()

	logrus.Infof("metrics server listen on %s", addr)

	if err := hs.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			logrus.Errorf("Listen failed %v", err)
		}
	}
}
