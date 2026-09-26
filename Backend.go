package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

/*
This file demonstrates a simple **sample backend client** implementation
that listens for messages published on a Redis channel ("Backend").
The purpose of this file is to serve as a reference for how a backend
service can subscribe to Redis events and process them further (e.g.,
forward to clients, trigger business logic, or store in a database).

Note: This is only an example to illustrate backend-side event handling.
In a real system, this logic can be extended with Propeller gRPC calls,
message routing, persistence, or any other domain-specific functionality.
*/

var ctx = context.Background()

func main() {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // change if Redis is on another host/port
		DB:   0,
	})

	// Subscribe to the "backend" channel
	subscriber := rdb.Subscribe(ctx, "Backend")

	// Get the channel to receive messages
	ch := subscriber.Channel()

	fmt.Println("Backend client subscribed to Redis channel 'backend'...")

	// Listen for messages
	for msg := range ch {
		fmt.Printf("Received message from Redis: %s\n", msg.Payload)

		// TODO: process event here (e.g. send to clients, store in DB, etc.)
	}
}
