package models

import (
	"gBalancer/logger"
	"testing"

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

	err = db.AutoMigrate(&RateLimits{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestCreateClient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	client := &RateLimits{ClientID: "client1", Capacity: 100, RatePerSec: 10}
	err = CreateClient(db, client)
	assert.NoError(t, err)

	var retrieved RateLimits
	err = db.First(&retrieved, "client_id = ?", "client1").Error
	assert.NoError(t, err)
	assert.Equal(t, client, &retrieved)
}

func TestUpdateAllClient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	client := &RateLimits{ClientID: "client1", Capacity: 100, RatePerSec: 10}
	CreateClient(db, client)

	err = UpdateAllClient(db, 20, 200)
	assert.NoError(t, err)

	var updated RateLimits
	err = db.First(&updated, "client_id = ?", "client1").Error
	assert.NoError(t, err)
	assert.Equal(t, 200, updated.Capacity)
	assert.Equal(t, 20, updated.RatePerSec)
}

func TestUpdateClient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	client := &RateLimits{ClientID: "client1", Capacity: 100, RatePerSec: 10}
	CreateClient(db, client)

	client.RatePerSec = 15
	err = UpdateClient(db, client)
	assert.NoError(t, err)

	var updated RateLimits
	err = db.First(&updated, "client_id = ?", "client1").Error
	assert.NoError(t, err)
	assert.Equal(t, 15, updated.RatePerSec)
}

func TestDeleteClient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	client := &RateLimits{ClientID: "client1", Capacity: 100, RatePerSec: 10}
	CreateClient(db, client)

	err = DeleteClient(db, "client1")
	assert.NoError(t, err)

	var deleted RateLimits
	err = db.First(&deleted, "client_id = ?", "client1").Error
	assert.Error(t, err)
}

func TestGetClient(t *testing.T) {
	db, err := setupTestDB()
	assert.NoError(t, err)

	client := &RateLimits{ClientID: "client1", Capacity: 100, RatePerSec: 10}
	CreateClient(db, client)

	retrieved, err := GetClient(db, "client1")
	assert.NoError(t, err)
	assert.Equal(t, client, retrieved)
}
