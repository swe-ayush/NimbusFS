package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/swe-ayush/nimbusfs/services"
)

var ctx = context.Background()

func main(){
	fmt.Println("Starting NimbusFS Storage Node agent...")

	// use redis ATOMIC INCR to name nodes
	rdb := redis.NewClient(&redis.Options{
								Addr: "localhost:6379",
								Password: "",
								DB: 0,
							})
	nodeNum, err := rdb.Incr(ctx, "nimbusfs:global:node_counter").Result()
	if err != nil {
		log.Fatalf("Redis failed : %v", err)
	}

	nodeID := fmt.Sprintf("node_%02d", nodeNum)
	fmt.Printf("Redis verified Identity, Storage Agent is running as %s\n", nodeID)

	engine := services.NewLocalStorageManager("vg_nimbus_global")
	diskPath := fmt.Sprintf("./storage_devices/%s.img", nodeID)
	mockDiskSize := uint64(2)
	err = engine.ExtendVolumeGroup(diskPath, uint64(mockDiskSize))

	if err != nil {
		log.Fatalf("Storage Hardware virtualization failed : %v", err)
	}

	fmt.Println("It's raining Gigabytes")
	fmt.Printf("Storage forecast : %s\n", "Nimbus_Unlimited")

	fmt.Scanln()
}