package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gBalancer/controllers"
	"gBalancer/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.RateLimits{})
	return db
}

func TestCreateClient(t *testing.T) {
	db := setupDatabase()
	handler := controllers.ClientHandler(db)

	client := models.RateLimits{ClientID: "test-client"}
	body, _ := json.Marshal(client)

	req, err := http.NewRequest(http.MethodPost, "/clients", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdClient models.RateLimits
	json.NewDecoder(rr.Body).Decode(&createdClient)
	assert.Equal(t, client.ClientID, createdClient.ClientID)
}

func TestGetClient(t *testing.T) {
	db := setupDatabase()
	handler := controllers.ClientHandler(db)

	client := models.RateLimits{ClientID: "test-client"}
	db.Create(&client)

	req, err := http.NewRequest(http.MethodGet, "/clients/test-client", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var retrievedClient models.RateLimits
	json.NewDecoder(rr.Body).Decode(&retrievedClient)
	assert.Equal(t, client.ClientID, retrievedClient.ClientID)
}

func TestUpdateClient(t *testing.T) {
	db := setupDatabase()
	handler := controllers.ClientHandler(db)

	client := models.RateLimits{ClientID: "test-client"}
	db.Create(&client)

	updatedClient := models.RateLimits{ClientID: "test-client"}
	body, _ := json.Marshal(updatedClient)

	req, err := http.NewRequest(http.MethodPut, "/clients/test-client", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestDeleteClient(t *testing.T) {
	db := setupDatabase()
	handler := controllers.ClientHandler(db)

	client := models.RateLimits{ClientID: "test-client"}
	db.Create(&client)

	req, err := http.NewRequest(http.MethodDelete, "/clients/test-client", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	var deletedClient models.RateLimits
	result := db.First(&deletedClient, "client_id = ?", "test-client")
	assert.Error(t, result.Error)
}
