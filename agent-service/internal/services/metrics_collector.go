package services

import (
	"context"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"hysteria2_microservices/agent-service/internal/config"
)

// MetricsCollectorImpl implements MetricsCollector interface
type MetricsCollectorImpl struct {
	logger          *logrus.Logger
	collectInterval time.Duration
	reportInterval  time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewMetricsCollector creates a new MetricsCollector
func NewMetricsCollector(cfg *config.Config, logger *logrus.Logger) MetricsCollector {
	return &MetricsCollectorImpl{
		logger:          logger,
		collectInterval: time.Duration(cfg.Metrics.CollectInterval) * time.Second,
		reportInterval:  time.Duration(cfg.Metrics.ReportInterval) * time.Second,
	}
}

// Collect collects current system metrics
func (mc *MetricsCollectorImpl) Collect() (map[string]interface{}, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"cpu_cores":    runtime.NumCPU(),
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"memory_total": m.TotalAlloc,
		"memory_sys":   m.Sys,
		"memory_gc":    m.NumGC,
		"timestamp":    time.Now().Unix(),
	}

	return metrics, nil
}

// StartCollection starts periodic metrics collection
func (mc *MetricsCollectorImpl) StartCollection(ctx context.Context) error {
	mc.ctx, mc.cancel = context.WithCancel(ctx)
	mc.logger.Info("Starting metrics collection")
	go mc.collectionLoop()
	return nil
}

// StopCollection stops metrics collection
func (mc *MetricsCollectorImpl) StopCollection() error {
	mc.logger.Info("Stopping metrics collection")
	if mc.cancel != nil {
		mc.cancel()
	}
	return nil
}

func (mc *MetricsCollectorImpl) collectionLoop() {
	ticker := time.NewTicker(mc.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics, err := mc.Collect()
			if err != nil {
				mc.logger.Errorf("Failed to collect metrics: %v", err)
				continue
			}
			mc.logger.Debugf("Collected metrics: %+v", metrics)
		case <-mc.ctx.Done():
			return
		}
	}
}
