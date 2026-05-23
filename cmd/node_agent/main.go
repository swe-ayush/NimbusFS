package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/swe-ayush/nimbusfs/pkg/pb"
	"github.com/swe-ayush/nimbusfs/services"
	"google.golang.org/grpc"
)

var ctx = context.Background()

func main(){
	fmt.Println("Starting NimbusFS Storage pool...")

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
	// TODO : as soon as redis increments, push a new disk to the pool, update the pool.
	nodeID := fmt.Sprintf("node_%02d", nodeNum)
	fmt.Printf("Redis verified Identity, Storage Agent is running as %s\n", nodeID)

	engine := services.NewLocalStorageManager("vg_nimbus_global")
	diskPath := fmt.Sprintf("./storage_devices/%s.img", nodeID)

	// baseline we can use as bigger disk not to choke our servers all of a sudden, demo-wise 2 is enough or even 1 is.
	mockDiskSize := 2.0
	err = engine.ExtendVolumeGroup(diskPath, mockDiskSize)

	if err != nil {
		log.Fatalf("Storage Hardware virtualization failed : %v", err)
	}

	fmt.Println("It's raining Gigabytes")
	fmt.Printf("Storage forecast : %s\n", "Nimbus_Unlimited")

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Unable to bind gRPC newtork layer to :50051. : %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStorageServiceServer(grpcServer, engine);

	go func() {
		fmt.Printf("[%s] bound to Global Pool. Listening on gRPC port :50051\n", nodeID)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("gRPC server encountered fatal execution failure: %v", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan 
	fmt.Printf("\n[%s] Detaching from global cluster pool and shutting down gracefully...\n", nodeID)
	grpcServer.GracefulStop()
	fmt.Println("Storage Node Agent completely stopped.")
}