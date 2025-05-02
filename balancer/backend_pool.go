package balancer

import (
	"fmt"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// BackendPool contains information about available backends
type BackendPool struct {
	backends     []*Backend
	current      uint64
	loadBalancer LoadBalancer
}

// Create new BackendPool based on url array
func NewBackendPool(urls []string, lb LoadBalancer) (*BackendPool, error) {
	var backendPool BackendPool
	backendPool.loadBalancer = lb

	for _, serverUrl := range urls {
		if err := AddBackendToPool(serverUrl, &backendPool); err != nil {
			return nil, fmt.Errorf("failed to add backend %s: %w", serverUrl, err)
		}
	}

	return &backendPool, nil
}

// Adds a backend to the backend pool
func (bp *BackendPool) AddBackend(backend *Backend) {
	bp.backends = append(bp.backends, backend)
}

// Atomically increments the counter and returns the index
func (bp *BackendPool) Next() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.backends)))
}

// Changes the availability status of the backend
func (bp *BackendPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range bp.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// Returns the next active peer to take connection
func (bp *BackendPool) GetNextPeer() *Backend {
	return bp.loadBalancer.GetNextBackend(bp)
}

// Checks the availability of the backend
func (bp *BackendPool) healthCheck() {
	for _, b := range bp.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		logrus.Infof("%s [%s]", b.URL, status)
	}
}

// isBackendAlive checks if the backend service at the given URL is reachable.
// It attempts to establish a TCP connection to the host specified in the URL.
// If the connection is successful within the specified timeout, it returns true.
// Otherwise, it logs an error and returns false.
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		logrus.Errorf("Site unreachable, error > %s", err)
		return false
	}
	defer conn.Close()
	return true
}

// HealthCheck periodically performs health checks on the backend services.
// It runs at the specified interval and logs the start and completion of each health check.
// The function listens for a signal on the stopChan to gracefully terminate the health check process.
func (bp *BackendPool) HealthCheck(interval time.Duration, stopChan <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			logrus.Info("Starting health check...")
			bp.healthCheck()
			logrus.Info("Health check completed")
		case <-stopChan:
			logrus.Info("Stopping health check...")
			return
		}
	}
}
