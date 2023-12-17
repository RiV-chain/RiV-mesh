//go:build !linux && !darwin && !windows && !openbsd && !freebsd
// +build !linux,!darwin,!windows,!openbsd,!freebsd

package config

// Sane defaults for the other platforms. The "default" options may be
// may be replaced by the running configuration.
func getDefaults() platformDefaultParameters {
	return platformDefaultParameters{

		// Configuration (used for meshctl)
		DefaultConfigFile: "/etc/mesh.conf",

		// Multicast interfaces
		DefaultMulticastInterfaces: []MulticastInterfaceConfig{
			{Regex: ".*", Beacon: true, Listen: true},
		},

		// TUN
		MaximumIfMTU:  65535,
		DefaultIfMTU:  65535,
		DefaultIfName: "none",
	}
}
