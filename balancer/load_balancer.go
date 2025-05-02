package balancer

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type LoadBalancer interface {
	GetNextBackend(bp *BackendPool) *Backend
}

type RoundRobin struct{}

func (rr *RoundRobin) GetNextBackend(bp *BackendPool) *Backend {
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

type Random struct{}

func (r *Random) GetNextBackend(bp *BackendPool) *Backend {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	aliveBackends := make([]*Backend, 0)
	for _, backend := range bp.backends {
		if backend.IsAlive() {
			aliveBackends = append(aliveBackends, backend)
		}
	}

	if len(aliveBackends) == 0 {
		return nil
	}

	randomIndex := random.Intn(len(aliveBackends))

	atomic.StoreUint64(&bp.current, uint64(randomIndex))

	return aliveBackends[randomIndex]
}
