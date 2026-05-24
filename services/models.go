package services

import (
	"context"
	"sync"

	"github.com/swe-ayush/nimbusfs/pkg/pb"
)

type PhysicalVolume struct {
	ID       string `json:"id"`
	FilePath string `json:"file_path"` 
	SizeGB   float64 `json:"size_gb"`
}

type VirtualLogicalVolume struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SizeGB float64 `json:"size_gb"`
	Path   string `json:"path"` 
}

type VirtualVolumeGroup struct {
	mu             sync.RWMutex
	Name           string                            `json:"name"`
	TotalCapacity  float64                            `json:"total_capacity_gb"`
	AllocatedSpace float64                            `json:"allocated_space_gb"`
	PhysicalDisks  map[string]PhysicalVolume         `json:"physical_disks"`
	LogicalVolumes map[string]*VirtualLogicalVolume  `json:"logical_volumes"`
}

// StorageManager defines what a single Storage Node Agent can do locally
type StorageManager interface {
	ExtendVolumeGroup(diskPath string, sizeGB float64) error
	CreateLogicalVolume(ctx context.Context, req *pb.CreateLogicalVolumeRequest) (*pb.CreateLogicalVolumeResponse, error)
	GetCapacity() (total float64, free float64)
}