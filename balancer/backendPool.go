package balancer

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
)

// BackendPool содержит информацию о доступных бэкендах
type BackendPool struct {
	backends     []*Backend
	current      uint64
	loadBalancer LoadBalancer
}

// Cоздает новый BackendPool на основе массива url
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

// Добавляет бэкенд в backend pool
func (bp *BackendPool) AddBackend(backend *Backend) {
	bp.backends = append(bp.backends, backend)
}

// Атомарно увеличивает счетчик и вовзращает индекс
func (bp *BackendPool) Next() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.backends)))
}

// Меняет статус доступности бэкенда
func (bp *BackendPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range bp.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

// возвращает следующий активный peer to take connection
func (bp *BackendPool) GetNextPeer() *Backend {
	return bp.loadBalancer.GetNextBackend(bp)
}

// Проверяет доступность бекенда
func (bp *BackendPool) healthCheck() {
	for _, b := range bp.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

// Обертка с задержкой над healthCheck
func (bp *BackendPool) HealthCheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			bp.healthCheck()
			log.Println("Health check completed")
		}
	}
}
