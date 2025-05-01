package balancer

import "sync/atomic"

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
