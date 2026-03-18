package worker

import (
	"context"
	"log/slog"
	"time"

	"insider-one-case/internal/config"
)

type IngestWorker struct {
	cfg     config.Config
	log     *slog.Logger
	batcher *Batcher
}

func NewIngestWorker(cfg config.Config, log *slog.Logger) *IngestWorker {
	return &IngestWorker{
		cfg:     cfg,
		log:     log,
		batcher: NewBatcher(cfg.WorkerBatchSize),
	}
}

func (w *IngestWorker) Start(ctx context.Context) {
	interval := time.Duration(w.cfg.WorkerFlushIntervalMS) * time.Millisecond
	if interval <= 0 {
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.log.Info("ingest worker started", "batch_size", w.cfg.WorkerBatchSize, "flush_interval", interval.String())

	for {
		select {
		case <-ctx.Done():
			w.log.Info("ingest worker stopped")
			return
		case <-ticker.C:
			// TODO: flush queued events to ClickHouse in batches.
			w.log.Debug("ingest worker tick", "pending_batch_items", w.batcher.Size())
		}
	}
}
