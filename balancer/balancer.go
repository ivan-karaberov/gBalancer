package balancer

import (
	"fmt"
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Attempts int = iota
	Retry
)

// Backend holds the data about a server
type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// Устанавливает статус доступности для бэкенда
func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

// Возвращает статус доступности для бэкенда
func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	defer b.mux.RUnlock()
	alive = b.Alive
	return
}

// BackendPool содержит информацию о доступных бэкендах
type BackendPool struct {
	backends []*Backend
	current  uint64
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
	next := bp.Next()
	l := len(bp.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(bp.backends)
		if bp.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&bp.current, uint64(idx))
			}
			return bp.backends[idx]
		}
	}
	return nil
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

// Cоздает новый BackendPool на основе массива url
func NewBackendPool(urls []string) (*BackendPool, error) {
	var backendPool BackendPool

	for _, serverUrl := range urls {
		if err := AddBackendToPool(serverUrl, &backendPool); err != nil {
			return nil, fmt.Errorf("failed to add backend %s: %w", serverUrl, err)
		}
	}

	return &backendPool, nil
}
