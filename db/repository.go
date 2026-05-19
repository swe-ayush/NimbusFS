package db

type MetadataRepository interface {
	SaveVolumeMetadata(volumeID string, nodeID string, sizeGB uint64) error
	GetVolumeLocation(volumeID string) (nodeID string, error error)
	UpdateNodeCapacity(nodeID string, freeSpaceGB uint64) error
}