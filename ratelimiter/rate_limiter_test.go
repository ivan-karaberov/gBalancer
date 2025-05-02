package ratelimiter

import (
	"testing"
	"time"

	"gBalancer/config"
	"gBalancer/logger"
	"gBalancer/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.CustomGormLogger(),
	})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.RateLimits{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(5, 1)

	for i := 0; i < 5; i++ {
		assert.True(t, tb.Allow())
	}

	assert.False(t, tb.Allow())

	time.Sleep(1 * time.Second)

	assert.True(t, tb.Allow())
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(5, 1)

	for i := 0; i < 5; i++ {
		tb.Allow()
	}

	assert.Equal(t, 0, tb.Tokens)

	time.Sleep(2 * time.Second)
	tb.Refill()

	assert.Equal(t, 2, tb.Tokens)
}

func TestTokenBucketManager_GetBucket(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	cfg := &config.Config{
		Capacity:   10,
		RatePerSec: 2,
	}
	manager := NewTokenBucketManager(db, cfg)

	clientID := "client1"

	bucket, err := manager.GetBucket(clientID, true)
	assert.NoError(t, err)
	assert.NotNil(t, bucket)
	assert.Equal(t, cfg.Capacity, bucket.Capacity)
	assert.Equal(t, cfg.RatePerSec, bucket.RatePerSec)

	bucket2, err := manager.GetBucket(clientID, false)
	assert.NoError(t, err)
	assert.Equal(t, bucket, bucket2)
}

func TestTokenBucketManager_GetBucket_NoCreate(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	cfg := &config.Config{
		Capacity:   10,
		RatePerSec: 2,
	}
	manager := NewTokenBucketManager(db, cfg)

	clientID := "nonexistent_client"

	bucket, err := manager.GetBucket(clientID, false)
	assert.Error(t, err)
	assert.Nil(t, bucket)
}
