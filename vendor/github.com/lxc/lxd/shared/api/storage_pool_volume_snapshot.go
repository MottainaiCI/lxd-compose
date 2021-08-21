package api

import "time"

// StorageVolumeSnapshotsPost represents the fields available for a new LXD storage volume snapshot
//
// swagger:model
//
// API extension: storage_api_volume_snapshots
type StorageVolumeSnapshotsPost struct {
	// Snapshot name
	// Example: snap0
	Name string `json:"name" yaml:"name"`

	// When the snapshot expires (gets auto-deleted)
	// Example: 2021-03-23T17:38:37.753398689-04:00
	//
	// API extension: custom_volume_snapshot_expiry
	ExpiresAt *time.Time `json:"expires_at" yaml:"expires_at"`
}

// StorageVolumeSnapshotPost represents the fields required to rename/move a LXD storage volume snapshot
//
// swagger:model
//
// API extension: storage_api_volume_snapshots
type StorageVolumeSnapshotPost struct {
	// New snapshot name
	// Example: snap1
	Name string `json:"name" yaml:"name"`
}

// StorageVolumeSnapshot represents a LXD storage volume snapshot
//
// swagger:model
//
// API extension: storage_api_volume_snapshots
type StorageVolumeSnapshot struct {
	StorageVolumeSnapshotPut `json:",inline" yaml:",inline"`

	// Snapshot name
	// Example: snap0
	Name string `json:"name" yaml:"name"`

	// Storage volume configuration map (refer to doc/storage.md)
	// Example: {"zfs.remove_snapshots": "true", "size": "50GiB"}
	Config map[string]string `json:"config" yaml:"config"`

	// The content type (filesystem or block)
	// Example: filesystem
	//
	// API extension: custom_block_volumes
	ContentType string `json:"content_type" yaml:"content_type"`
}

// StorageVolumeSnapshotPut represents the modifiable fields of a LXD storage volume
//
// swagger:model
//
// API extension: storage_api_volume_snapshots
type StorageVolumeSnapshotPut struct {
	// Description of the storage volume
	// Example: My custom volume
	Description string `json:"description" yaml:"description"`

	// When the snapshot expires (gets auto-deleted)
	// Example: 2021-03-23T17:38:37.753398689-04:00
	//
	// API extension: custom_volume_snapshot_expiry
	ExpiresAt *time.Time `json:"expires_at" yaml:"expires_at"`
}
