package coordinator

import (
	"context"
	"fmt"
	"ncobase/core/workflow/service"
	"sync"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"

	"github.com/hashicorp/consul/api"
)

// Config represents coordinator configuration
type Config struct {
	NodeID              string
	Namespace           string
	LockTTL             time.Duration
	RetryInterval       time.Duration
	MaxRetries          int
	HealthCheckInterval time.Duration
	DeregisterTimeout   time.Duration
}

// DefaultConfig returns default coordinator configuration
func DefaultConfig() *Config {
	return &Config{
		LockTTL:             5 * time.Minute,
		RetryInterval:       5 * time.Second,
		MaxRetries:          3,
		HealthCheckInterval: 30 * time.Second,
		DeregisterTimeout:   5 * time.Minute,
	}
}

// Coordinator represents a workflow coordinator
type Coordinator struct {
	services  *service.Service
	em        ext.ManagerInterface
	consul    *api.Client
	config    *Config
	logger    logger.Logger
	locks     map[string]*api.Lock
	locksMu   sync.RWMutex
	isLeader  bool
	leaderCh  chan bool
	leaderKey string
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewCoordinator creates a new workflow coordinator
func NewCoordinator(svc *service.Service, em ext.ManagerInterface, consulClient *api.Client, cfg *Config) *Coordinator {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Coordinator{
		services:  svc,
		em:        em,
		consul:    consulClient,
		config:    cfg,
		locks:     make(map[string]*api.Lock),
		leaderCh:  make(chan bool, 1),
		leaderKey: fmt.Sprintf("%s/workflow/leader", cfg.Namespace),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the coordinator
func (c *Coordinator) Start() error {
	if err := c.registerService(); err != nil {
		return fmt.Errorf("register service: %w", err)
	}
	c.wg.Add(1)
	go c.runLeaderElection()
	c.wg.Add(1)
	go c.runHealthCheck()
	return nil
}

// Stop stops the coordinator
func (c *Coordinator) Stop() error {
	c.cancel()
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Minute):
		return fmt.Errorf("timeout waiting for shutdown")
	}
	c.releaseAllLocks()
	return c.deregisterService()
}

// AcquireLock acquires a distributed lock
func (c *Coordinator) AcquireLock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("%s/workflow/locks/%s", c.config.Namespace, key)
	opts := &api.LockOptions{
		Key:         lockKey,
		SessionName: fmt.Sprintf("workflow-lock-%s-%s", c.config.NodeID, key),
		SessionTTL:  c.config.LockTTL.String(),
		LockTryOnce: true,
	}
	lock, err := c.consul.LockOpts(opts)
	if err != nil {
		return fmt.Errorf("create lock: %w", err)
	}
	stopCh := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			close(stopCh)
		case <-c.ctx.Done():
			close(stopCh)
		}
	}()
	leaderCh, err := lock.Lock(stopCh)
	if err != nil || leaderCh == nil {
		return fmt.Errorf("failed to acquire lock")
	}
	c.locksMu.Lock()
	c.locks[key] = lock
	c.locksMu.Unlock()
	return nil
}

// ReleaseLock releases a distributed lock
func (c *Coordinator) ReleaseLock(key string) error {
	c.locksMu.Lock()
	defer c.locksMu.Unlock()
	if lock, exists := c.locks[key]; exists {
		if err := lock.Unlock(); err != nil {
			return fmt.Errorf("release lock: %w", err)
		}
		delete(c.locks, key)
	}
	return nil
}

// IsLeader returns whether this node is the leader
func (c *Coordinator) IsLeader() bool {
	return c.isLeader
}

// WaitForLeader waits until this node becomes leader or context is cancelled
func (c *Coordinator) WaitForLeader(ctx context.Context) error {
	if c.isLeader {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.leaderCh:
		return nil
	}
}

// Internal methods

func (c *Coordinator) runLeaderElection() {
	defer c.wg.Done()
	opts := &api.LockOptions{
		Key:         c.leaderKey,
		SessionName: fmt.Sprintf("workflow-leader-%s", c.config.NodeID),
		SessionTTL:  c.config.LockTTL.String(),
	}
	lock, err := c.consul.LockOpts(opts)
	if err != nil {
		c.logger.Errorf(c.ctx, "create leader lock failed: %v", err)
		return
	}
	for retryCount := 0; ; retryCount = 0 {
		select {
		case <-c.ctx.Done():
			return
		default:
			leaderCh, err := lock.Lock(nil)
			if err != nil {
				if retryCount++; retryCount > c.config.MaxRetries {
					c.logger.Errorf(c.ctx, "leader election failed after %d retries", retryCount)
					return
				}
				time.Sleep(c.config.RetryInterval)
				continue
			}
			c.isLeader = true
			c.leaderCh <- true
			c.em.PublishEvent("workflow.leader.elected", map[string]any{"node_id": c.config.NodeID})
			select {
			case <-c.ctx.Done():
				err := lock.Unlock()
				if err != nil {
					c.logger.Errorf(c.ctx, "release leader lock failed: %v", err)
					return
				}
				return
			case <-leaderCh:
				c.isLeader = false
				c.em.PublishEvent("workflow.leader.lost", map[string]any{"node_id": c.config.NodeID})
			}
		}
	}
}

func (c *Coordinator) runHealthCheck() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.config.HealthCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.updateHealthCheck(); err != nil {
				c.logger.Errorf(c.ctx, "health check failed: %v", err)
			}
		}
	}
}

func (c *Coordinator) releaseAllLocks() {
	c.locksMu.Lock()
	defer c.locksMu.Unlock()
	for key, lock := range c.locks {
		if err := lock.Unlock(); err != nil {
			c.logger.Errorf(c.ctx, "release lock %s failed: %v", key, err)
		}
		delete(c.locks, key)
	}
}

func (c *Coordinator) registerService() error {
	s := &api.AgentServiceRegistration{
		ID:   fmt.Sprintf("workflow-%s", c.config.NodeID),
		Name: "workflow",
		Tags: []string{"workflow", c.config.Namespace},
		Check: &api.AgentServiceCheck{
			TTL:                            c.config.HealthCheckInterval.String(),
			DeregisterCriticalServiceAfter: c.config.DeregisterTimeout.String(),
		},
	}
	return c.consul.Agent().ServiceRegister(s)
}

func (c *Coordinator) deregisterService() error {
	return c.consul.Agent().ServiceDeregister(fmt.Sprintf("workflow-%s", c.config.NodeID))
}

func (c *Coordinator) updateHealthCheck() error {
	return c.consul.Agent().UpdateTTL(fmt.Sprintf("service:workflow-%s", c.config.NodeID), "healthy", "passing")
}
