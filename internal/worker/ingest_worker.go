package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"insider-one-case/internal/config"
	"insider-one-case/internal/model"
)

type IngestWorker struct {
	cfg     config.Config
	log     *slog.Logger
	batcher *Batcher
	queue   chan model.EventIngestRequest
}

func NewIngestWorker(cfg config.Config, log *slog.Logger) *IngestWorker {
	queueSize := cfg.WorkerBatchSize * 10
	if queueSize <= 0 {
		queueSize = 1000
	}

	return &IngestWorker{
		cfg:     cfg,
		log:     log,
		batcher: NewBatcher(cfg.WorkerBatchSize),
		queue:   make(chan model.EventIngestRequest, queueSize),
	}
}

// Enqueue adds an event to the in-memory queue for async processing.
// Returns an error if the queue is full.
func (w *IngestWorker) Enqueue(event model.EventIngestRequest) error {
	select {
	case w.queue <- event:
		return nil
	default:
		return errors.New("event queue is full")
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
			w.flush()
		}
	}
}

func (w *IngestWorker) flush() {
	var batch []model.EventIngestRequest

drain:
	for {
		select {
		case event := <-w.queue:
			batch = append(batch, event)
			if len(batch) >= w.batcher.size {
				break drain
			}
		default:
			break drain
		}
	}

	if len(batch) == 0 {
		return
	}

	w.log.Debug("flushing event batch", "count", len(batch))
	// TODO: persist batch to ClickHouse via repository.
}
