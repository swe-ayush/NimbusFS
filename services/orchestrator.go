package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/swe-ayush/nimbusfs/pkg/pb"
)

type ClusterOrchestrator struct {
	storageClient pb.StorageServiceClient
}

func NewOrchestrator(client pb.StorageServiceClient) *ClusterOrchestrator {
	return &ClusterOrchestrator{
		storageClient: client,
	}
}

// Background job to monitor storage pool metrics
func (o *ClusterOrchestrator) StartCapacityMonitor(ctx context.Context, checkInterval time.Duration) {
	ticker := time.NewTicker(checkInterval)
	log.Println("[ORCHESTRATOR] Background capacity pool monitor successfully started...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				o.checkPoolCapacity(ctx)
			}
		}
	}()
}

func (o *ClusterOrchestrator) checkPoolCapacity(ctx context.Context) {
	stats, err := o.storageClient.GetStats(ctx, &pb.GetsStatsRequest{})
	if err != nil {
		log.Printf("[ORCHESTRATOR ERROR] Failed to fetch pool metrics: %v", err)
		return
	}

	// Bootstrap edge case: if the global pool returns zero capacity, kickstart it with an initial slice
	if stats.TotalSpaceGb == 0 {
		log.Printf("[ORCHESTRATOR] Pool reports 0 GB capacity. Initializing first expansion track...")
		o.triggerPoolExpansion(ctx, 2)
		return
	}

	freePercent := (stats.FreeSpaceGb / stats.TotalSpaceGb) * 100
	log.Printf("[POOL UPDATE] Volume Group: %s | Total: %.2f GB | Free: %.2f GB (%.2f%% free)", 
		stats.NodeId, stats.TotalSpaceGb, stats.FreeSpaceGb, freePercent)

	if freePercent <= 20.0 {
		log.Printf("[🚨 THRESHOLD BREACHED] Pool %s running low! Initiating auto-provisioning...", stats.NodeId)
		go o.triggerPoolExpansion(ctx, 2) 
	}
}

func (o *ClusterOrchestrator) triggerPoolExpansion(ctx context.Context, expandSizeGb float64) {
	uniqueChunkID := fmt.Sprintf("chunk_auto_%d", time.Now().Unix())
	req := &pb.CreateVolumeRequest{
		VolumeId: uniqueChunkID,
		SizeGb:   expandSizeGb,
	}

	resp, err := o.storageClient.CreateVolume(ctx, req)
	if err != nil {
		log.Printf("[EXPANSION FAILED] Could not scale pool over the wire: %v", err)
		return
	}

	if resp.Success {
		log.Printf("[EXPANSION SUCCESS] Global pool scaled up successfully! Target chunk path: %s", resp.Path)
	} else {
		log.Printf("[EXPANSION REJECTED] Storage engine failed operation: %s", resp.Message)
	}
}