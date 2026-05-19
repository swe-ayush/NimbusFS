package services

import (
	"sync"
)

type PhysicalVolume struct {
	ID       string `json:"id"`
	FilePath string `json:"file_path"` 
	SizeGB   uint64 `json:"size_gb"`
}

type VirtualLogicalVolume struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	SizeGB uint64 `json:"size_gb"`
	Path   string `json:"path"` 
}

type VirtualVolumeGroup struct {
	mu             sync.RWMutex
	Name           string                            `json:"name"`
	TotalCapacity  uint64                            `json:"total_capacity_gb"`
	AllocatedSpace uint64                            `json:"allocated_space_gb"`
	PhysicalDisks  map[string]PhysicalVolume         `json:"physical_disks"`
	LogicalVolumes map[string]*VirtualLogicalVolume  `json:"logical_volumes"`
}

// StorageManager defines what a single Storage Node Agent can do locally
type StorageManager interface {
	ExtendVolumeGroup(diskID string, sizeGB uint64) error
	CreateLogicalVolume(volumeID string, sizeGB uint64) (*VirtualLogicalVolume, error)
	GetCapacity() (total uint64, free uint64)
}