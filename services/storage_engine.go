package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/swe-ayush/nimbusfs/pkg/pb"
)

type LocalStorageManager struct{
	pb.UnimplementedStorageServiceServer
	VG *VirtualVolumeGroup
}

func NewLocalStorageManager(name string) *LocalStorageManager{
	return &LocalStorageManager{
		VG : &VirtualVolumeGroup{
			Name: name,
			PhysicalDisks: make(map[string]PhysicalVolume),
			LogicalVolumes: make(map[string]*VirtualLogicalVolume),
		},
	}
}

func (m *LocalStorageManager) CreateVolume(ctx context.Context, req *pb.CreateVolumeRequest) (*pb.CreateVolumeResponse, error) {
	chunkID := req.GetVolumeId()
	sizeGB := req.GetSizeGb()
	chunkPath := fmt.Sprintf("./storage_devices/%s.img", chunkID)
	err := m.ExtendVolumeGroup(chunkPath, sizeGB)
	if err != nil {
		return &pb.CreateVolumeResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to accomodate chunk:%s in the pool : %v", chunkID, err),
		}, nil
	}

	return &pb.CreateVolumeResponse{
		Success: true,
		Message: fmt.Sprintf("Capacity Expansion Successful by volume : %s for %.2f GB", chunkID, sizeGB),
		Path: chunkPath,
	}, nil
}

func (m *LocalStorageManager) GetStats(ctx context.Context, req *pb.GetsStatsRequest) (*pb.GetsStatsResponse, error) {
	total, free := m.GetCapacity()
	return &pb.GetsStatsResponse{
		NodeId: m.VG.Name,
		TotalSpaceGb: total,
		FreeSpaceGb: free,
	}, nil
}

func (m *LocalStorageManager) ExtendVolumeGroup(diskPath string, sizeGB float64) error{
	m.VG.mu.Lock()
	defer m.VG.mu.Unlock()

	file, err := os.Create(diskPath)
	if err != nil {
		log.Fatalf("Failed to create mock disk %s : %v", diskPath, err)
	}

	defer file.Close()
	sizeInBytes := uint64(sizeGB * 1024 * 1024 * 1024)
	err = file.Truncate(int64(sizeInBytes))
	if err != nil {
		return fmt.Errorf("failed to allocate sparse space via truncate: %w", err)
	}

	id := filepath.Base(diskPath)
	m.VG.PhysicalDisks[id] = PhysicalVolume{
		ID: id,
		FilePath: diskPath,
		SizeGB: sizeGB,
	}

	m.VG.TotalCapacity += sizeGB

	return nil
}

func (m *LocalStorageManager) CreateLogicalVolume(volumeID string, sizeGB float64) (*VirtualLogicalVolume, error){
	m.VG.mu.Lock()
	defer m.VG.mu.Unlock()
	// logical section
	return nil,nil
}

func (m *LocalStorageManager) GetCapacity() (total float64, free float64){
	m.VG.mu.RLock()
	defer m.VG.mu.RUnlock()

	freeSpace := m.VG.TotalCapacity - m.VG.AllocatedSpace
	return m.VG.TotalCapacity, freeSpace
}