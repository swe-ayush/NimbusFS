package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/swe-ayush/nimbusfs/pkg/pb"
	"github.com/swe-ayush/nimbusfs/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
type UserVolumeRequest struct {
	VolumeID   string  `json:"volume_id"`
	VolumeName string  `json:"volume_name"`
	SizeGB     float64 `json:"size_gb"`
}

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
	// select {}

	http.HandleFunc("/api/volumes", handleCreateLogicalVolume(storageClient))
	log.Println("Control Plane API Gateway online on HTTP port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("HTTP Gateway failed to launch: %v", err)
	}

}

func handleCreateLogicalVolume(client pb.StorageServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}

		var req UserVolumeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload format", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		stats, err := client.GetStats(ctx, &pb.GetsStatsRequest{})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "failed",
				"error":  fmt.Sprintf("Failed to fetch node metrics for validation: %v", err),
			})
			return
		}

		log.Printf("[API HIT] Allocation requested for '%s' (%s) requiring %.2f GB", req.VolumeID, req.VolumeName, req.SizeGB)

		if req.SizeGB > stats.GetFreeSpaceGb() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInsufficientStorage)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "denied",
				"error":  fmt.Sprintf("Requested allocation (%.2f GB) exceeds available pool limits (%.2f GB available)", req.SizeGB, stats.GetFreeSpaceGb()),
			})
			return
		}

		nodeResponse, err := client.CreateLogicalVolume(ctx, &pb.CreateLogicalVolumeRequest{
			VolumeId:   req.VolumeID,
			VolumeName: req.VolumeName,
			SizeGb:     req.SizeGB,
		})

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "failed",
				"error":  fmt.Sprintf("RPC network communication failure during provisioning: %v", err),
			})
			return
		}

		if !nodeResponse.GetSuccess() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "failed",
				"error":  nodeResponse.GetMessage(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success_allocated",
			"id":          req.VolumeID,
			"name":        req.VolumeName,
			"allocated_gb": req.SizeGB,
			"device_path": nodeResponse.GetPath(),
			"node_message": nodeResponse.GetMessage(),
		})
	}
}
