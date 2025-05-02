package middleware

import (
	"gBalancer/errors"
	"gBalancer/ratelimiter"
	"gBalancer/utils"
	"net/http"
)

// Creates middleware for token management.
func RateLimitMiddleware(manager *ratelimiter.TokenBucketManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := utils.GetIP(r)

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
