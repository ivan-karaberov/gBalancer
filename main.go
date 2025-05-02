package main

import (
	"fmt"
	"net/http"
	"time"

	"gBalancer/balancer"
	"gBalancer/config"
	"gBalancer/controllers"
	"gBalancer/logger"
	"gBalancer/middleware"
	"gBalancer/models"
	"gBalancer/ratelimiter"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logger.CustomFormatter{})

	config, err := config.ReadConfig("config.json")
	if err != nil {
		logrus.Fatalf("failed read balancer config > %s", err)
	}

	backendPool, err := balancer.NewBackendPool(config.Backends, &balancer.RoundRobin{})
	if err != nil {
		logrus.Fatalf("failed create balancer > %s", err)
	}

	db := models.NewDBConnection()

	models.UpdateAllClient(db, config.RatePerSec, config.Capacity)

	bucketManager := ratelimiter.NewTokenBucketManager(db, config)

	mux := http.NewServeMux()
	mux.Handle("/clients/", http.HandlerFunc(controllers.ClientHandler(db)))
	mux.Handle("/", middleware.RateLimitMiddleware(bucketManager)(http.HandlerFunc(backendPool.HandleRequest)))

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}

	stopChan := make(chan struct{})
	go backendPool.HealthCheck(time.Minute*2, stopChan)
	defer close(stopChan)

	logrus.Infof("Load Balancer started at :%d", config.Port)
	if err := server.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}
