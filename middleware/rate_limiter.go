package middleware

import (
	"gBalancer/errors"
	"gBalancer/ratelimiter"
	"net"
	"net/http"
	"strings"
)

// Creates middleware for token management.
func RateLimitMiddleware(manager *ratelimiter.TokenBucketManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getIP(r)

			bucket, err := manager.GetBucket(clientIP, true)
			if err != nil {
				errors.APIError(w, errors.ErrClientNotFound)
				return
			}

			if !bucket.Allow() {
				errors.APIError(w, errors.ErrRateLimitExceeded)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Retrieves the client's IP address from the HTTP request.
func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	ip = r.RemoteAddr
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	return host
}
