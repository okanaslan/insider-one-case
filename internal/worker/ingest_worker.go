package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"insider-one-case/internal/config"
	"insider-one-case/internal/model"
)

var ErrEnqueueTimeout = errors.New("enqueue timed out")

type EventBatchWriter interface {
	InsertEventsBatch(ctx context.Context, events []model.EventIngestRequest) error
}

type IngestWorker struct {
	cfg    config.Config
	log    *slog.Logger
	queue  chan model.EventIngestRequest
	writer EventBatchWriter
}

func NewIngestWorker(cfg config.Config, log *slog.Logger, writer EventBatchWriter) *IngestWorker {
	queueSize := cfg.IngestQueueBufferSize
	if queueSize <= 0 {
		queueSize = 10000
	}

	return &IngestWorker{
		cfg:    cfg,
		log:    log,
		queue:  make(chan model.EventIngestRequest, queueSize),
		writer: writer,
	}
}

// Enqueue adds an event to the in-memory queue for async processing.

// Returns ErrEnqueueTimeout if the queue cannot accept the event within
// the configured enqueue timeout window.
func (w *IngestWorker) Enqueue(ctx context.Context, event model.EventIngestRequest) error {
	timeoutMS := w.cfg.IngestEnqueueTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 25
	}

	enqueueCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	select {
	case w.queue <- event:
		return nil
	case <-enqueueCtx.Done():
		if errors.Is(enqueueCtx.Err(), context.DeadlineExceeded) {
			return ErrEnqueueTimeout
		}
		return enqueueCtx.Err()
	}
}

func (w *IngestWorker) Start(ctx context.Context) {
	batchSize := w.cfg.WorkerBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	interval := time.Duration(w.cfg.WorkerFlushIntervalMS) * time.Millisecond
	if interval <= 0 {
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	w.log.Info("ingest worker started",
		"batch_size", batchSize,
		"flush_interval", interval.String(),
		"queue_buffer", cap(w.queue),
	)

	batch := make([]model.EventIngestRequest, 0, batchSize)

	for {
		select {
		case event := <-w.queue:
			batch = append(batch, event)
			if len(batch) >= batchSize {
				if err := w.flushBatch(ctx, batch); err != nil {
					// Keep behavior explicit and simple: drop failed batch after logging.
					w.log.Error("failed to flush full batch", "count", len(batch), "error", err)
				}
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) == 0 {
				continue
			}
			if err := w.flushBatch(ctx, batch); err != nil {
				w.log.Error("failed to flush batch on ticker", "count", len(batch), "error", err)
			}
			batch = batch[:0]

		case <-ctx.Done():
			batch = w.drainQueuedEvents(batch)
			if len(batch) > 0 {
				w.log.Info("worker shutdown flush attempt", "pending_count", len(batch))
				if err := w.flushBatch(context.Background(), batch); err != nil {
					w.log.Error("failed to flush batch during shutdown", "count", len(batch), "error", err)
				}
			}
			w.log.Info("ingest worker stopped")
			return
		}
	}
}

func (w *IngestWorker) flushBatch(ctx context.Context, batch []model.EventIngestRequest) error {
	if len(batch) == 0 {
		return nil
	}

	if w.writer == nil {
		w.log.Debug("event writer not configured, skipping batch flush", "count", len(batch))
		return nil
	}

	if err := w.writer.InsertEventsBatch(ctx, batch); err != nil {
		return err
	}

	w.log.Info("event batch flushed", "count", len(batch))
	return nil
}

func (w *IngestWorker) drainQueuedEvents(batch []model.EventIngestRequest) []model.EventIngestRequest {
	for {
		select {
		case event := <-w.queue:
			batch = append(batch, event)
		default:
			return batch
		}
	}
}
