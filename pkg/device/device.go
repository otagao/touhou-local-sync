// Package device handles device identification using hostname and MAC address.
package device

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/smelt02/touhou-local-sync/pkg/utils"
)

// GetDeviceID generates a unique device ID based on hostname and primary MAC address.
// Returns: device_id (first 12 chars of SHA256(hostname+mac)), full hash, hostname, error
func GetDeviceID() (id string, hash string, hostname string, err error) {
	// Get hostname
	hostname, err = os.Hostname()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get hostname: %w", err)
	}

	// Get primary MAC address
	mac, err := getPrimaryMAC()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get MAC address: %w", err)
	}

	// Calculate hash: SHA256(hostname + mac)
	combined := hostname + mac
	fullHash := utils.CalculateStringHash(combined)

	// Device ID is first 12 characters of hash
	if len(fullHash) < 12 {
		return "", "", "", fmt.Errorf("hash too short: %s", fullHash)
	}
	deviceID := fullHash[:12]

	// Return full hash with "sha256:" prefix for storage
	hashWithPrefix := "sha256:" + fullHash

	return deviceID, hashWithPrefix, hostname, nil
}

// getPrimaryMAC returns the MAC address of the first non-loopback network interface.
// Returns the MAC address as a string (e.g., "00:11:22:33:44:55").
func getPrimaryMAC() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// Skip loopback and interfaces without MAC address
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		// Skip interfaces that are down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Found a valid interface
		mac := iface.HardwareAddr.String()
		if mac != "" {
			return strings.ToLower(mac), nil
		}
	}

	return "", fmt.Errorf("no valid network interface found")
}
