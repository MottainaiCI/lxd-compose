package api

// NetworkPeersPost represents the fields of a new LXD network peering
//
// swagger:model
//
// API extension: network_peer
type NetworkPeersPost struct {
	NetworkPeerPut `yaml:",inline"`

	// Name of the peer
	// Example: project1-network1
	Name string `json:"name" yaml:"name"`

	// Name of the target project
	// Example: project1
	TargetProject string `json:"target_project" yaml:"target_project"`

	// Name of the target network
	// Example: network1
	TargetNetwork string `json:"target_network" yaml:"target_network"`
}

// NetworkPeerPut represents the modifiable fields of a LXD network peering
//
// swagger:model
//
// API extension: network_peer
type NetworkPeerPut struct {
	// Description of the peer
	// Example: Peering with network1 in project1
	Description string `json:"description" yaml:"description"`

	// Peer configuration map (refer to doc/network-peers.md)
	// Example: {"user.mykey": "foo"}
	Config map[string]string `json:"config" yaml:"config"`
}

// NetworkPeer used for displaying a LXD network peering.
//
// swagger:model
//
// API extension: network_forward
type NetworkPeer struct {
	NetworkPeerPut `yaml:",inline"`

	// Name of the peer
	// Read only: true
	// Example: project1-network1
	Name string `json:"name" yaml:"name"`

	// Name of the target project
	// Read only: true
	// Example: project1
	TargetProject string `json:"target_project" yaml:"target_project"`

	// Name of the target network
	// Read only: true
	// Example: network1
	TargetNetwork string `json:"target_network" yaml:"target_network"`

	// The state of the peering
	// Read only: true
	// Example: Pending
	Status string `json:"status" yaml:"status"`

	// List of URLs of objects using this network peering
	// Read only: true
	// Example: ["/1.0/network-acls/test", "/1.0/network-acls/foo"]
	UsedBy []string `json:"used_by" yaml:"used_by"`
}

// Etag returns the values used for etag generation.
func (p *NetworkPeer) Etag() []any {
	return []any{p.Name, p.Description, p.Config}
}

// Writable converts a full NetworkPeer struct into a NetworkPeerPut struct (filters read-only fields).
func (p *NetworkPeer) Writable() NetworkPeerPut {
	return p.NetworkPeerPut
}
