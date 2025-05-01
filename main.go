package main

import (
	"fmt"
	"log"
	"net/http"

	"gBalancer/balancer"
	"gBalancer/config"
)

func main() {
	config, err := config.ReadConfig("config.json")
	if err != nil {
		log.Fatalf("failed read balancer config > %s", err)
	}

	backendPool, err := balancer.NewBackendPool(config.Backends, &balancer.RoundRobin{})
	if err != nil {
		log.Fatalf("failed create balancer > %s", err)
	}

	server := http.Server{
		Addr:    fmt.Sprintf("%s:%d", "localhost", config.Port),
		Handler: http.HandlerFunc(backendPool.HandleRequest),
	}

	go backendPool.HealthCheck()

	log.Printf("Load Balancer started at :%d\n", config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
