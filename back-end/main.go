package main

import (
	"log"

	"sensor-api/messaging"
	"sensor-api/router"
)

func main() {
	pool, err := messaging.GetPool()
	if err != nil {
		log.Printf("[main] Aviso: pool RabbitMQ não inicializado: %v", err)
	} else {
		defer pool.Close()
	}

	r := router.Setup()
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}