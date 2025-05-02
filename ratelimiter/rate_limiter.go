package ratelimiter

import (
	"gBalancer/config"
	"gBalancer/models"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Structure for managing tokens for the client.
type TokenBucket struct {
	Capacity   int        // Maximum number of tokens
	RatePerSec int        // Rate of token replenishment per second
	Tokens     int        // Current number of tokens
	LastRefill time.Time  // Time of the last refill
	mu         sync.Mutex // Mutex for thread safety
}

// Create a Token Bucket with a given capacity and replenishment rate.
func NewTokenBucket(capacity, ratePerSec int) *TokenBucket {
	return &TokenBucket{
		Capacity:   capacity,
		RatePerSec: ratePerSec,
		Tokens:     capacity, // Initially, all tokens are available
		LastRefill: time.Now(),
	}
}

// Refill replenishes tokens in a bucket.
func (tb *TokenBucket) Refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.LastRefill).Seconds()
	tb.LastRefill = now

	// Calculate the number of tokens that need to be added
	newTokens := int(elapsed * float64(tb.RatePerSec))
	if newTokens > 0 {
		tb.Tokens += newTokens
		if tb.Tokens > tb.Capacity {
			tb.Tokens = tb.Capacity
		}
	}
}

// Checks whether the request can be fulfilled and updates the state of the tokens.
func (tb *TokenBucket) Allow() bool {
	tb.Refill()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.Tokens > 0 {
		tb.Tokens--
		return true
	}
	return false
}

// Manages tokens for all clients.
type TokenBucketManager struct {
	buckets map[string]*TokenBucket // Dictionary for storing client tokens
	mu      sync.Mutex              // Mutex for thread safety
	cfg     *config.Config
	DB      *gorm.DB
}

func NewTokenBucketManager(db *gorm.DB, cfg *config.Config) *TokenBucketManager {
	return &TokenBucketManager{
		buckets: make(map[string]*TokenBucket),
		cfg:     cfg,
		DB:      db,
	}
}

// Obtains a token bucket for the client, creating it if necessary.
func (manager *TokenBucketManager) GetBucket(clientID string, createIfNotExists bool) (*TokenBucket, error) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if bucket, exists := manager.buckets[clientID]; exists {
		return bucket, nil
	}

	rateLimits, err := models.GetClient(manager.DB, clientID)
	if err != nil {
		if createIfNotExists {
			rateLimits = &models.RateLimits{
				ClientID:   clientID,
				Capacity:   manager.cfg.Capacity,
				RatePerSec: manager.cfg.RatePerSec,
			}
			models.CreateClient(manager.DB, rateLimits)
		} else {
			return nil, err
		}
	}

	bucket := NewTokenBucket(rateLimits.Capacity, rateLimits.RatePerSec)
	manager.buckets[clientID] = bucket
	return bucket, nil
}
