package balancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
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
