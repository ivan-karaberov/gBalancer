package balancer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func (bp *BackendPool) LoadBalancer(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return

	}
	peer := bp.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func AddBackendToPool(serverUrl string, bp *BackendPool) error {
	parsedUrl, err := url.Parse(serverUrl)
	if err != nil {
		log.Fatal(err)
	}

	proxy, err := createReverseProxy(parsedUrl, bp)
	if err != nil {
		return fmt.Errorf("failed to create reverse proxy for %s: %w", serverUrl, err)
	}

	bp.AddBackend(&Backend{
		URL:          parsedUrl,
		Alive:        true,
		ReverseProxy: proxy,
	})
	log.Printf("Configured server: %s\n", serverUrl)
	return nil
}

func createReverseProxy(serverUrl *url.URL, pool *BackendPool) (*httputil.ReverseProxy, error) {
	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s] %s\n", serverUrl, e.Error())
		retries := GetRetryFromContext(request)
		if retries < 3 {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}
		// after 3 retries, mark this backend as down
		pool.MarkBackendStatus(serverUrl, false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		pool.LoadBalancer(writer, request.WithContext(ctx))
	}
	return proxy, nil
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}
