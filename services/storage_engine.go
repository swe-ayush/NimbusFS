package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LocalStorageManager struct{
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

func (m *LocalStorageManager) ExtendVolumeGroup(diskPath string, sizeGB uint64) error{
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

func (m *LocalStorageManager) CreateLogicalVolume(volumeID string, sizeGB uint64) (*VirtualLogicalVolume, error){
	m.VG.mu.Lock()
	defer m.VG.mu.Unlock()
	// logical section
	return nil,nil
}

func (m *LocalStorageManager) GetCapacity() (total uint64, free uint64){
	m.VG.mu.RLock()
	defer m.VG.mu.RUnlock()

	freeSpace := m.VG.TotalCapacity - m.VG.AllocatedSpace
	return m.VG.TotalCapacity, freeSpace
}