package balancer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

// Определяем пользовательский тип для ключа Retry
type contextKeyRetry struct{}

// Определяем пользовательский тип для ключа Attempts
type contextKeyAttempts struct{}

var (
	Retry    contextKeyRetry    = contextKeyRetry{}
	Attempts contextKeyAttempts = contextKeyAttempts{}
)

// HandleRequest processes incoming HTTP requests for the BackendPool.
// It checks the number of attempts made by the client and routes the request
// to the next available peer for handling.
func (bp *BackendPool) HandleRequest(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		logrus.Warningf("%s(%s) Max attempts reached, terminating", r.RemoteAddr, r.URL.Path)
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

// Adds a new backend server to the specified BackendPool.
func AddBackendToPool(serverUrl string, bp *BackendPool) error {
	parsedUrl, err := url.Parse(serverUrl)
	if err != nil {
		return fmt.Errorf("failed parse server URL %s: %w", serverUrl, err)
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
	logrus.Infof("Configured server: %s", serverUrl)
	return nil
}

// Initializes a new reverse proxy for the specified server URL
func createReverseProxy(serverUrl *url.URL, pool *BackendPool) (*httputil.ReverseProxy, error) {
	proxy := httputil.NewSingleHostReverseProxy(serverUrl)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		logrus.Errorf("[%s] %s", serverUrl, e.Error())
		retries := GetRetryFromContext(request)
		if retries < 3 {
			<-time.After(10 * time.Millisecond)

			ctx := context.WithValue(request.Context(), Retry, retries+1)
			if ctx.Err() == nil {
				proxy.ServeHTTP(writer, request.WithContext(ctx))
			}
			return
		}
		// after 3 retries, mark this backend as down
		pool.MarkBackendStatus(serverUrl, false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := GetAttemptsFromContext(request)
		logrus.Warningf("%s(%s) Attempting retry %d", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)
		pool.HandleRequest(writer, request.WithContext(ctx))
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
