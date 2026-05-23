package main

import (
	"context"
	"log"
	"time"

	"github.com/swe-ayush/nimbusfs/pkg/pb"
	"github.com/swe-ayush/nimbusfs/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log.Println("Orchestrating Nimbus...")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil{
		log.Fatalf("Orchestrator failed to connect to Storage pool : %v", err)
	}

	defer conn.Close()

	storageClient := pb.NewStorageServiceClient(conn)
	orchestratorMonitor := services.NewOrchestrator(storageClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Orchestrator successfully attached to storage node pipeline.")
	
	orchestratorMonitor.StartCapacityMonitor(ctx, 5*time.Second)

	// keeps daemon running
	select {}

}
