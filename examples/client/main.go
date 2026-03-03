package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/priyanshu360/cachemesh/client"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	c := client.New("localhost:8080")
	defer c.Close()

	ctx := context.Background()

	if err := c.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}
	fmt.Println("Ping: OK")

	user := User{Name: "John", Email: "john@example.com"}
	err := c.Set(ctx, "user:1", user, time.Hour)
	if err != nil {
		log.Fatalf("Failed to set: %v", err)
	}
	fmt.Println("Set: OK")

	var retrieved User
	err = c.GetTo(ctx, "user:1", &retrieved)
	if err != nil {
		log.Fatalf("Failed to get: %v", err)
	}
	fmt.Printf("Get: %+v\n", retrieved)

	exists, err := c.Exist(ctx, "user:1")
	if err != nil {
		log.Fatalf("Failed to check exist: %v", err)
	}
	fmt.Printf("Exist: %v\n", exists)

	deleted, err := c.Delete(ctx, "user:1")
	if err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}
	fmt.Printf("Delete: %v\n", deleted)

	exists, _ = c.Exist(ctx, "user:1")
	fmt.Printf("Exist after delete: %v\n", exists)

	fmt.Println("\n--- Cluster Mode ---")

	cluster := client.NewCluster([]string{
		"localhost:8080",
		"localhost:8081",
		"localhost:8082",
	})
	defer cluster.Close()

	keys := []string{"user:1", "user:2", "user:3", "product:100", "order:500"}
	for _, key := range keys {
		err := cluster.Set(ctx, key, fmt.Sprintf("value-%s", key), time.Hour)
		if err != nil {
			log.Printf("Failed to set %s: %v", key, err)
			continue
		}
		val, _ := cluster.Get(ctx, key)
		fmt.Printf("Key %s -> %v\n", key, val)
	}

	cluster.Invalidate(ctx, "user:1")
	fmt.Println("Invalidated user:1 across all nodes")
}
