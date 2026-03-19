package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/config"
	"insider-one-case/internal/model"
)

type errBatchWriter struct {
	err error
}

func (e *errBatchWriter) InsertEventsBatch(ctx context.Context, events []model.EventIngestRequest) error {
	_ = ctx
	_ = events
	return e.err
}

type fakeBatchWriter struct {
	mu      sync.Mutex
	calls   int
	batches [][]model.EventIngestRequest
}

func (f *fakeBatchWriter) InsertEventsBatch(ctx context.Context, events []model.EventIngestRequest) error {
	_ = ctx

	f.mu.Lock()
	defer f.mu.Unlock()

	copied := make([]model.EventIngestRequest, len(events))
	copy(copied, events)

	f.calls++
	f.batches = append(f.batches, copied)
	return nil
}

func (f *fakeBatchWriter) snapshot() (int, [][]model.EventIngestRequest) {
	f.mu.Lock()
	defer f.mu.Unlock()

	copiedBatches := make([][]model.EventIngestRequest, len(f.batches))
	for index, batch := range f.batches {
		batchCopy := make([]model.EventIngestRequest, len(batch))
		copy(batchCopy, batch)
		copiedBatches[index] = batchCopy
	}

	return f.calls, copiedBatches
}

func testEvent(name string) model.EventIngestRequest {
	return model.EventIngestRequest{
		EventName:  name,
		Channel:    "mobile",
		CampaignID: "cmp_123",
		UserID:     "user_456",
		Timestamp:  1710000000,
		Tags:       []string{"promo"},
		Metadata:   map[string]any{"amount": 120},
	}
}

func TestIngestWorkerFlushesWhenBatchSizeReached(t *testing.T) {
	writer := &fakeBatchWriter{}
	worker := NewIngestWorker(config.Config{
		WorkerBatchSize:       2,
		WorkerFlushIntervalMS: 60000,
		IngestQueueBufferSize: 10,
	}, slog.Default(), writer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	require.NoError(t, worker.Enqueue(context.Background(), testEvent("purchase")))
	require.NoError(t, worker.Enqueue(context.Background(), testEvent("signup")))

	require.Eventually(t, func() bool {
		calls, batches := writer.snapshot()
		return calls == 1 && len(batches) == 1 && len(batches[0]) == 2
	}, time.Second, 10*time.Millisecond)
}

func TestIngestWorkerFlushesOnTimer(t *testing.T) {
	writer := &fakeBatchWriter{}
	worker := NewIngestWorker(config.Config{
		WorkerBatchSize:       10,
		WorkerFlushIntervalMS: 20,
		IngestQueueBufferSize: 10,
	}, slog.Default(), writer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	require.NoError(t, worker.Enqueue(context.Background(), testEvent("purchase")))

	require.Eventually(t, func() bool {
		calls, batches := writer.snapshot()
		return calls == 1 && len(batches) == 1 && len(batches[0]) == 1
	}, time.Second, 10*time.Millisecond)
}

func TestIngestWorkerFlushesQueuedEventsOnShutdown(t *testing.T) {
	writer := &fakeBatchWriter{}
	worker := NewIngestWorker(config.Config{
		WorkerBatchSize:       10,
		WorkerFlushIntervalMS: 60000,
		IngestQueueBufferSize: 10,
	}, slog.Default(), writer)

	require.NoError(t, worker.Enqueue(context.Background(), testEvent("purchase")))
	require.NoError(t, worker.Enqueue(context.Background(), testEvent("signup")))

	ctx, cancel := context.WithCancel(context.Background())
	go worker.Start(ctx)
	cancel()

	require.Eventually(t, func() bool {
		calls, batches := writer.snapshot()
		return calls == 1 && len(batches) == 1 && len(batches[0]) == 2
	}, time.Second, 10*time.Millisecond)
}

func TestIngestWorkerEnqueueReturnsTimeoutWhenQueueIsFull(t *testing.T) {
	worker := NewIngestWorker(config.Config{
		WorkerBatchSize:        100,
		WorkerFlushIntervalMS:  60000,
		IngestQueueBufferSize:  1,
		IngestEnqueueTimeoutMS: 20,
	}, slog.Default(), &errBatchWriter{err: errors.New("insert failed")})

	require.NoError(t, worker.Enqueue(context.Background(), testEvent("purchase")))

	err := worker.Enqueue(context.Background(), testEvent("signup"))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrEnqueueTimeout)
}

func TestIngestWorkerEnqueueReturnsContextCanceled(t *testing.T) {
	worker := NewIngestWorker(config.Config{
		IngestQueueBufferSize:  1,
		IngestEnqueueTimeoutMS: 100,
	}, slog.Default(), &errBatchWriter{err: errors.New("insert failed")})

	require.NoError(t, worker.Enqueue(context.Background(), testEvent("purchase")))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := worker.Enqueue(ctx, testEvent("signup"))
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
