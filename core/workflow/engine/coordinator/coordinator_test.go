package coordinator

import (
	"context"
	"testing"
	"time"

	ext "github.com/ncobase/ncore/ext/types"

	"ncobase/core/workflow/service"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

// Helper functions for setup and teardown
func setupTestCoordinator() (*Coordinator, func()) {
	// Mock dependencies
	svc := &service.Service{}
	em, _ := ext.NewManager(nil)
	client, _ := api.NewClient(api.DefaultConfig())
	cfg := DefaultConfig()

	// Create coordinator
	coord := NewCoordinator(svc, em, client, cfg)
	return coord, func() {
		coord.Stop()
	}
}

func TestNewCoordinator_DefaultConfig(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	assert.Equal(t, time.Minute*5, coord.config.LockTTL)
	assert.Equal(t, time.Second*5, coord.config.RetryInterval)
	assert.Equal(t, 3, coord.config.MaxRetries)
}

func TestStartStopCoordinator(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err, "Coordinator should start without error")

	err = coord.Stop()
	assert.NoError(t, err, "Coordinator should stop without error")
}

func TestAcquireReleaseLock(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	lockKey := "test-lock"
	err = coord.AcquireLock(ctx, lockKey)
	assert.NoError(t, err, "Acquiring lock should succeed")

	err = coord.ReleaseLock(lockKey)
	assert.NoError(t, err, "Releasing lock should succeed")
}

func TestAcquireLock_ReacquireAfterRelease(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	lockKey := "test-lock"
	err = coord.AcquireLock(ctx, lockKey)
	assert.NoError(t, err, "Acquiring lock should succeed")

	err = coord.ReleaseLock(lockKey)
	assert.NoError(t, err, "Releasing lock should succeed")

	// Attempt to acquire the lock again
	err = coord.AcquireLock(ctx, lockKey)
	assert.NoError(t, err, "Reacquiring lock should succeed after release")
}

func TestIsLeader(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	// Simulate coordinator as leader
	coord.isLeader = true
	assert.True(t, coord.IsLeader(), "Coordinator should be leader")

	// Revoke leadership
	coord.isLeader = false
	assert.False(t, coord.IsLeader(), "Coordinator should not be leader after revoking")
}

func TestWaitForLeader(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	// Initially, should not be leader
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	go func() {
		time.Sleep(time.Second * 2)
		coord.leaderCh <- true // Simulate leadership acquisition
	}()

	err = coord.WaitForLeader(ctx)
	assert.NoError(t, err, "Should not return error when waiting for leader")
}

func TestHealthCheck(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	// Mock health check behavior
	err = coord.updateHealthCheck()
	assert.NoError(t, err, "Health check should succeed")
}

func TestReleaseAllLocks(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	ctx := context.Background()
	lockKeys := []string{"lock-1", "lock-2", "lock-3"}

	for _, key := range lockKeys {
		err := coord.AcquireLock(ctx, key)
		assert.NoError(t, err, "Acquiring lock should succeed")
	}

	coord.releaseAllLocks()

	for _, key := range lockKeys {
		assert.NotContains(t, coord.locks, key, "Lock should be removed from coordinator")
	}
}

func TestDeregisterService(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	// Test deregister service
	err = coord.deregisterService()
	assert.NoError(t, err, "Service should deregister without error")
}

func TestUpdateHealthCheck(t *testing.T) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	err := coord.Start()
	assert.NoError(t, err)

	err = coord.updateHealthCheck()
	assert.NoError(t, err, "Health check update should succeed")
}

func BenchmarkAcquireLock(b *testing.B) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	coord.Start()
	defer coord.Stop()

	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		lockKey := "benchmark-lock"
		coord.AcquireLock(ctx, lockKey)
		coord.ReleaseLock(lockKey)
	}
}

func BenchmarkReleaseLock(b *testing.B) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	coord.Start()
	defer coord.Stop()

	ctx := context.Background()
	lockKey := "benchmark-lock"
	coord.AcquireLock(ctx, lockKey)

	b.ResetTimer() // Reset timer to only measure the release

	for i := 0; i < b.N; i++ {
		coord.ReleaseLock(lockKey)
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	coord, teardown := setupTestCoordinator()
	defer teardown()

	coord.Start()
	defer coord.Stop()

	for i := 0; i < b.N; i++ {
		coord.updateHealthCheck()
	}
}
