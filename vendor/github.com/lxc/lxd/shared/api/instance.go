package api

import (
	"time"
)

// InstanceType represents the type if instance being returned or requested via the API.
type InstanceType string

// InstanceTypeAny defines the instance type value for requesting any instance type.
const InstanceTypeAny = InstanceType("")

// InstanceTypeContainer defines the instance type value for a container.
const InstanceTypeContainer = InstanceType("container")

// InstanceTypeVM defines the instance type value for a virtual-machine.
const InstanceTypeVM = InstanceType("virtual-machine")

// InstancesPost represents the fields available for a new LXD instance.
//
// API extension: instances
type InstancesPost struct {
	InstancePut `yaml:",inline"`

	Name         string         `json:"name" yaml:"name"`
	Source       InstanceSource `json:"source" yaml:"source"`
	InstanceType string         `json:"instance_type" yaml:"instance_type"`
	Type         InstanceType   `json:"type" yaml:"type"`
}

// InstancesPut represents the fields available for a mass update.
//
// API extension: instance_bulk_state_change
type InstancesPut struct {
	State *InstanceStatePut `json:"state" yaml:"state"`
}

// InstancePost represents the fields required to rename/move a LXD instance.
//
// swagger:model
//
// API extension: instances
type InstancePost struct {
	// New name for the instance
	// Example: bar
	Name string `json:"name" yaml:"name"`

	// Whether the instance is being migrated to another server
	// Example: false
	Migration bool `json:"migration" yaml:"migration"`

	// Whether to perform a live migration (migration only)
	// Example: false
	Live bool `json:"live" yaml:"live"`

	// Whether snapshots should be discarded (migration only)
	// Example: false
	InstanceOnly bool `json:"instance_only" yaml:"instance_only"`

	// Whether snapshots should be discarded (migration only, deprecated, use instance_only)
	// Example: false
	ContainerOnly bool `json:"container_only" yaml:"container_only"` // Deprecated, use InstanceOnly.

	// Target for the migration, will use pull mode if not set (migration only)
	Target *InstancePostTarget `json:"target" yaml:"target"`

	// Target pool for local cross-pool move
	// Example: baz
	//
	// API extension: instance_pool_move
	Pool string `json:"pool" yaml:"pool"`
}

// InstancePostTarget represents the migration target host and operation.
//
// swagger:model
//
// API extension: instances
type InstancePostTarget struct {
	// The certificate of the migration target
	// Example: X509 PEM certificate
	Certificate string `json:"certificate" yaml:"certificate"`

	// The operation URL on the remote target
	// Example: https://1.2.3.4:8443/1.0/operations/5e8e1638-5345-4c2d-bac9-2c79c8577292
	Operation string `json:"operation,omitempty" yaml:"operation,omitempty"`

	// Migration websockets credentials
	// Example: {"migration": "random-string", "criu": "random-string"}
	Websockets map[string]string `json:"secrets,omitempty" yaml:"secrets,omitempty"`
}

// InstancePut represents the modifiable fields of a LXD instance.
//
// swagger:model
//
// API extension: instances
type InstancePut struct {
	// Architecture name
	// Example: x86_64
	Architecture string `json:"architecture" yaml:"architecture"`

	// Instance configuration (see doc/instances.md)
	// Example: {"security.nesting": "true"}
	Config map[string]string `json:"config" yaml:"config"`

	// Instance devices (see doc/instances.md)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	Devices map[string]map[string]string `json:"devices" yaml:"devices"`

	// Whether the instance is ephemeral (deleted on shutdown)
	// Example: false
	Ephemeral bool `json:"ephemeral" yaml:"ephemeral"`

	// List of profiles applied to the instance
	// Example: ["default"]
	Profiles []string `json:"profiles" yaml:"profiles"`

	// If set, instance will be restored to the provided snapshot name
	// Example: snap0
	Restore string `json:"restore,omitempty" yaml:"restore,omitempty"`

	// Whether the instance currently has saved state on disk
	// Example: false
	Stateful bool `json:"stateful" yaml:"stateful"`

	// Instance description
	// Example: My test instance
	Description string `json:"description" yaml:"description"`
}

// Instance represents a LXD instance.
//
// swagger:model
//
// API extension: instances
type Instance struct {
	InstancePut `yaml:",inline"`

	// Instance creation timestamp
	// Example: 2021-03-23T20:00:00-04:00
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`

	// Expanded configuration (all profiles and local config merged)
	// Example: {"security.nesting": "true"}
	ExpandedConfig map[string]string `json:"expanded_config" yaml:"expanded_config"`

	// Expanded devices (all profiles and local devices merged)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	ExpandedDevices map[string]map[string]string `json:"expanded_devices" yaml:"expanded_devices"`

	// Instance name
	// Example: foo
	Name string `json:"name" yaml:"name"`

	// Instance status (see instance_state)
	// Example: Running
	Status string `json:"status" yaml:"status"`

	// Instance status code (see instance_state)
	// Example: 101
	StatusCode StatusCode `json:"status_code" yaml:"status_code"`

	// Last start timestamp
	// Example: 2021-03-23T20:00:00-04:00
	LastUsedAt time.Time `json:"last_used_at" yaml:"last_used_at"`

	// What cluster member this instance is located on
	// Example: lxd01
	Location string `json:"location" yaml:"location"`

	// The type of instance (container or virtual-machine)
	// Example: container
	Type string `json:"type" yaml:"type"`
}

// InstanceFull is a combination of Instance, InstanceBackup, InstanceState and InstanceSnapshot.
//
// API extension: instances
type InstanceFull struct {
	Instance `yaml:",inline"`

	Backups   []InstanceBackup   `json:"backups" yaml:"backups"`
	State     *InstanceState     `json:"state" yaml:"state"`
	Snapshots []InstanceSnapshot `json:"snapshots" yaml:"snapshots"`
}

// Writable converts a full Instance struct into a InstancePut struct (filters read-only fields).
//
// API extension: instances
func (c *Instance) Writable() InstancePut {
	return c.InstancePut
}

// IsActive checks whether the instance state indicates the instance is active.
//
// API extension: instances
func (c Instance) IsActive() bool {
	switch c.StatusCode {
	case Stopped:
		return false
	case Error:
		return false
	default:
		return true
	}
}

// InstanceSource represents the creation source for a new instance.
//
// API extension: instances
type InstanceSource struct {
	Type          string            `json:"type" yaml:"type"`
	Certificate   string            `json:"certificate" yaml:"certificate"`
	Alias         string            `json:"alias,omitempty" yaml:"alias,omitempty"`
	Fingerprint   string            `json:"fingerprint,omitempty" yaml:"fingerprint,omitempty"`
	Properties    map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
	Server        string            `json:"server,omitempty" yaml:"server,omitempty"`
	Secret        string            `json:"secret,omitempty" yaml:"secret,omitempty"`
	Protocol      string            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	BaseImage     string            `json:"base-image,omitempty" yaml:"base-image,omitempty"`
	Mode          string            `json:"mode,omitempty" yaml:"mode,omitempty"`
	Operation     string            `json:"operation,omitempty" yaml:"operation,omitempty"`
	Websockets    map[string]string `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Source        string            `json:"source,omitempty" yaml:"source,omitempty"`
	Live          bool              `json:"live,omitempty" yaml:"live,omitempty"`
	InstanceOnly  bool              `json:"instance_only,omitempty" yaml:"instance_only,omitempty"`
	ContainerOnly bool              `json:"container_only,omitempty" yaml:"container_only,omitempty"` // Deprecated, use InstanceOnly.
	Refresh       bool              `json:"refresh,omitempty" yaml:"refresh,omitempty"`
	Project       string            `json:"project,omitempty" yaml:"project,omitempty"`
}
