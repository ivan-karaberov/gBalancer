package controllers

import (
	"encoding/json"
	"gBalancer/errors"
	"gBalancer/models"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handles HTTP requests for clients.
func ClientHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetClient(db, w, r)
		case http.MethodPost:
			CreateClient(db, w, r)
		case http.MethodPut:
			UpdateClient(db, w, r)
		case http.MethodDelete:
			DeleteClient(db, w, r)
		default:
			errors.APIError(w, errors.ErrMethodNotAllowed)
			return
		}
	}
}

// Creates a new client based on the data provided in the request body.
func CreateClient(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	client := new(models.RateLimits)
	if err := json.NewDecoder(r.Body).Decode(client); err != nil {
		logrus.Errorf("JSON Decode error > %s", err.Error())
		errors.APIError(w, errors.ErrBadRequestBody)
		return
	}

	if client.ClientID == "" {
		errors.APIError(w, errors.ErrBadRequestBody)
		return
	}

	if err := models.CreateClient(db, client); err != nil {
		logrus.Errorf("Failed to create client: %s", err.Error())
		errors.APIError(w, errors.ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

// Returns information about a client by their identifier (clientID).
func GetClient(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) < 3 {
		errors.APIError(w, errors.ErrClientNotFound)
		return
	}

	clientID := splitPath[2]
	client, err := models.GetClient(db, clientID)
	if err != nil {
		logrus.Errorf("Error retrieving client > %s", err.Error())
		errors.APIError(w, errors.ErrClientNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(client)
}

// Updates client information based on data from the request body.
func UpdateClient(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	updClient := new(models.RateLimits)
	if err := json.NewDecoder(r.Body).Decode(updClient); err != nil {
		logrus.Errorf("JSON Decode error > %s", err.Error())
		errors.APIError(w, errors.ErrBadRequestBody)
		return
	}

	if updClient.ClientID == "" {
		errors.APIError(w, errors.ErrClientNotFound)
		return
	}

	err := models.UpdateClient(db, updClient)
	if err != nil {
		logrus.Errorf("Failed update client > %s", err.Error())
		errors.APIError(w, errors.ErrClientNotUpdated)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Removes a client by their identifier (clientID) provided in the URL.
func DeleteClient(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) < 3 {
		errors.APIError(w, errors.ErrClientNotFound)
		return
	}

	clientID := splitPath[2]
	err := models.DeleteClient(db, clientID)
	if err != nil {
		errors.APIError(w, errors.ErrClientNotDeleted)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
