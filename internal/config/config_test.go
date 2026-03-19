package config

import "testing"

func TestConfigNormalizedAppliesDefaultsForNonPositiveValues(t *testing.T) {
	t.Parallel()

	normalized := (Config{}).Normalized()

	if normalized.WorkerBatchSize != DefaultWorkerBatchSize {
		t.Fatalf("expected default worker batch size %d, got %d", DefaultWorkerBatchSize, normalized.WorkerBatchSize)
	}
	if normalized.WorkerFlushIntervalMS != DefaultWorkerFlushIntervalMS {
		t.Fatalf("expected default worker flush interval %d, got %d", DefaultWorkerFlushIntervalMS, normalized.WorkerFlushIntervalMS)
	}
	if normalized.IngestQueueBufferSize != DefaultIngestQueueBufferSize {
		t.Fatalf("expected default queue buffer size %d, got %d", DefaultIngestQueueBufferSize, normalized.IngestQueueBufferSize)
	}
	if normalized.IngestEnqueueTimeoutMS != DefaultIngestEnqueueTimeoutMS {
		t.Fatalf("expected default enqueue timeout %d, got %d", DefaultIngestEnqueueTimeoutMS, normalized.IngestEnqueueTimeoutMS)
	}
}

func TestConfigNormalizedClampsLargeValues(t *testing.T) {
	t.Parallel()

	normalized := (Config{
		WorkerBatchSize:        MaxWorkerBatchSize + 1,
		WorkerFlushIntervalMS:  MaxWorkerFlushIntervalMS + 1,
		IngestQueueBufferSize:  MaxIngestQueueBufferSize + 1,
		IngestEnqueueTimeoutMS: MaxIngestEnqueueTimeoutMS + 1,
	}).Normalized()

	if normalized.WorkerBatchSize != MaxWorkerBatchSize {
		t.Fatalf("expected clamped worker batch size %d, got %d", MaxWorkerBatchSize, normalized.WorkerBatchSize)
	}
	if normalized.WorkerFlushIntervalMS != MaxWorkerFlushIntervalMS {
		t.Fatalf("expected clamped worker flush interval %d, got %d", MaxWorkerFlushIntervalMS, normalized.WorkerFlushIntervalMS)
	}
	if normalized.IngestQueueBufferSize != MaxIngestQueueBufferSize {
		t.Fatalf("expected clamped queue buffer size %d, got %d", MaxIngestQueueBufferSize, normalized.IngestQueueBufferSize)
	}
	if normalized.IngestEnqueueTimeoutMS != MaxIngestEnqueueTimeoutMS {
		t.Fatalf("expected clamped enqueue timeout %d, got %d", MaxIngestEnqueueTimeoutMS, normalized.IngestEnqueueTimeoutMS)
	}
}
