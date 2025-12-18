// Package snapshotter provides a background worker manager for generating portfolio snapshots.
package snapshotter

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

var (
	jobsEnqueued = promauto.NewCounter(prometheus.CounterOpts{
		Name: "snapshot_jobs_enqueued_total",
		Help: "The total number of snapshot jobs added to the queue",
	})
	jobsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snapshot_jobs_processed_total",
		Help: "The total number of processed snapshots",
	}, []string{"status"}) // status = "success" or "error"

	queueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snapshot_queue_depth",
		Help: "Current number of jobs waiting in the buffer",
	})

	processingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "snapshot_processing_duration_seconds",
		Help:    "Time taken to generate a snapshot",
		Buckets: prometheus.DefBuckets,
	})
)

// Job represents a unit of work for the snapshotter
type Job struct {
	UserID     string
	JobID      string
	RetryCount int
}

// WorkerConfig holds command-line configuration for the snapshot worker
type WorkerConfig struct {
	PoolSize   int
	MaxRetries int
	RateLimit  float64 // Jobs per second
	Burst      int
}

// LockTTL is the configurable "Max Time" a snapshot generation is allowed to take
const LockTTL = 5 * time.Minute

// LockEntry tracks the active lock for a user
type LockEntry struct {
	StartTime time.Time
	JobID     string
}

// Manager orchestrates the snapshot generation workers
type Manager struct {
	queue   chan Job
	limiter *rate.Limiter
	config  WorkerConfig
	// Tracks users currently in the queue or being processed
	mu       sync.Mutex
	inFlight map[string]LockEntry
}

// NewManager creates a new snapshot manager
func NewManager(cfg WorkerConfig) *Manager {
	return &Manager{
		queue:    make(chan Job, 1000),
		limiter:  rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.Burst),
		config:   cfg,
		inFlight: make(map[string]LockEntry),
	}
}

// Start initiates the worker pool
func (m *Manager) Start(ctx context.Context, processor func(string) error) {
	for i := 0; i < m.config.PoolSize; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-m.queue:
					fmt.Printf("***Processing snapshot for user %s with job ID %s\n", job.UserID, job.JobID)

					// Wrap processing in a function to enforce defer scope
					func() {
						// ✅ SAFE: Only delete if WE own the lock
						defer m.safeUnlock(job.UserID, job.JobID)

						queueSize.Dec() // Decrement gauge as job is picked up

						// Respect rate limits
						if err := m.limiter.Wait(ctx); err != nil {
							return
						}

						start := time.Now()
						err := processor(job.UserID)

						// Record Metrics
						processingDuration.Observe(time.Since(start).Seconds())

						if err != nil {
							jobsProcessed.WithLabelValues("error").Inc()
							m.handleFailure(job, err)
						} else {
							jobsProcessed.WithLabelValues("success").Inc()
						}
					}()

				}
			}
		}()
	}
}

// Trigger queues a snapshot job for a user
func (m *Manager) Trigger(userID string) {
	fmt.Printf("***Triggering snapshot for user %s\n", userID)

	m.mu.Lock()
	entry, exists := m.inFlight[userID]

	// 1. Check for Active Lock
	if exists {
		if time.Since(entry.StartTime) < LockTTL {
			// Job is running and valid. Skip Trigger.
			m.mu.Unlock()
			return
		}
		// Else: Lock exists but is EXPIRED. We will overwrite it (Steal Lock).
		fmt.Printf("⚠️ Stale lock detected for user %s. Stealing lock.\n", userID)
	}

	// 2. Create New Lock (Lease)
	jobID := uuid.New().String()
	m.inFlight[userID] = LockEntry{
		StartTime: time.Now(),
		JobID:     jobID,
	}
	m.mu.Unlock()

	select {
	case m.queue <- Job{UserID: userID, JobID: jobID, RetryCount: 0}: // Init RetryCount
		jobsEnqueued.Inc()
		queueSize.Inc()
	default:
		// Queue full - maybe increment a "dropped_jobs" counter here

		// If queue is full, we must remove from inFlight
		// so it can be retried later
		m.safeUnlock(userID, jobID)
	}
}

func (m *Manager) handleFailure(job Job, err error) {
	if job.RetryCount < m.config.MaxRetries {
		job.RetryCount++
		// Exponential backoff: 2^count * 1s
		delay := time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second

		time.AfterFunc(delay, func() {
			m.queue <- job
		})
	} else {
		// Log to monitoring (e.g., Sentry/Prometheus)
		fmt.Printf("Critical: User %s failed after %d retries\n", job.UserID, job.RetryCount)
	}
}

// Helper to ensure we only delete OUR lock
func (m *Manager) safeUnlock(userID, jobID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry, exists := m.inFlight[userID]; exists {
		// Only delete if the JobID matches.
		// If IDs don't match, it means the lock was already stolen by a newer job.
		if entry.JobID == jobID {
			delete(m.inFlight, userID)
		}
	}
}
