package lxd

import (
	"fmt"
	"net/url"

	"github.com/lxc/lxd/shared/api"
)

// GetNetworkPeerNames returns a list of network peer names.
func (r *ProtocolLXD) GetNetworkPeerNames(networkName string) ([]string, error) {
	if !r.HasExtension("network_peer") {
		return nil, fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	// Fetch the raw URL values.
	urls := []string{}
	baseURL := fmt.Sprintf("/networks/%s/peers", url.PathEscape(networkName))
	_, err := r.queryStruct("GET", baseURL, nil, "", &urls)
	if err != nil {
		return nil, err
	}

	// Parse it.
	return urlsToResourceNames(baseURL, urls...)
}

// GetNetworkPeers returns a list of network peer structs.
func (r *ProtocolLXD) GetNetworkPeers(networkName string) ([]api.NetworkPeer, error) {
	if !r.HasExtension("network_peer") {
		return nil, fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	peers := []api.NetworkPeer{}

	// Fetch the raw value.
	_, err := r.queryStruct("GET", fmt.Sprintf("/networks/%s/peers?recursion=1", url.PathEscape(networkName)), nil, "", &peers)
	if err != nil {
		return nil, err
	}

	return peers, nil
}

// GetNetworkPeer returns a network peer entry for the provided network and peer name.
func (r *ProtocolLXD) GetNetworkPeer(networkName string, peerName string) (*api.NetworkPeer, string, error) {
	if !r.HasExtension("network_peer") {
		return nil, "", fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	peer := api.NetworkPeer{}

	// Fetch the raw value.
	etag, err := r.queryStruct("GET", fmt.Sprintf("/networks/%s/peers/%s", url.PathEscape(networkName), url.PathEscape(peerName)), nil, "", &peer)
	if err != nil {
		return nil, "", err
	}

	return &peer, etag, nil
}

// CreateNetworkPeer defines a new network peer using the provided struct.
// Returns true if the peer connection has been mutually created. Returns false if peering has been only initiated.
func (r *ProtocolLXD) CreateNetworkPeer(networkName string, peer api.NetworkPeersPost) error {
	if !r.HasExtension("network_peer") {
		return fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	// Send the request.
	_, _, err := r.query("POST", fmt.Sprintf("/networks/%s/peers", url.PathEscape(networkName)), peer, "")
	if err != nil {
		return err
	}

	return nil
}

// UpdateNetworkPeer updates the network peer to match the provided struct.
func (r *ProtocolLXD) UpdateNetworkPeer(networkName string, peerName string, peer api.NetworkPeerPut, ETag string) error {
	if !r.HasExtension("network_peer") {
		return fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	// Send the request.
	_, _, err := r.query("PUT", fmt.Sprintf("/networks/%s/peers/%s", url.PathEscape(networkName), url.PathEscape(peerName)), peer, ETag)
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetworkPeer deletes an existing network peer.
func (r *ProtocolLXD) DeleteNetworkPeer(networkName string, peerName string) error {
	if !r.HasExtension("network_peer") {
		return fmt.Errorf(`The server is missing the required "network_peer" API extension`)
	}

	// Send the request.
	_, _, err := r.query("DELETE", fmt.Sprintf("/networks/%s/peers/%s", url.PathEscape(networkName), url.PathEscape(peerName)), nil, "")
	if err != nil {
		return err
	}

	return nil
}
