package models

import (
	"gorm.io/gorm"
)

// Represents the rate limit data for a client, such as maximum capacity and requests per second.
type RateLimits struct {
	ClientID   string `json:"client_id" gorm:"primary_key;"` // Client identifier (primary key).
	Capacity   int    `json:"capacity" gorm:"not null;"`     // Maximum capacity for the client (e.g., maximum number of requests).
	RatePerSec int    `json:"rate_per_sec" gorm:"not null;"` // Maximum number of requests per second for the client.
}

// Adds a new client to the rate_limits table.
func CreateClient(db *gorm.DB, rl *RateLimits) error {
	return db.Create(&rl).Error
}

// Updated all client's in the rate_limits table
func UpdateAllClient(db *gorm.DB, rate int, capacity int) error {
	updates := RateLimits{
		Capacity:   capacity,
		RatePerSec: rate,
	}
	return db.Model(&RateLimits{}).Where("1 = 1").Updates(updates).Error
}

// Updates the client's data in the rate_limits table by their identifier.
func UpdateClient(db *gorm.DB, rl *RateLimits) error {
	return db.Model(&RateLimits{}).Where("client_id = ?", rl.ClientID).Updates(rl).Error
}

// Removes a client from the rate_limits table by their identifier.
func DeleteClient(db *gorm.DB, clientID string) error {
	return db.Where("client_id = ?", clientID).Delete(&RateLimits{}).Error
}

// Retrieves the client's data from the rate_limits table by their identifier.
func GetClient(db *gorm.DB, clientID string) (*RateLimits, error) {
	var client RateLimits
	err := db.Where("client_id = ?", clientID).First(&client).Error
	return &client, err
}
