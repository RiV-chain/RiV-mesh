package mobile

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"

	"github.com/gologme/log"

	"github.com/RiV-chain/RiV-mesh/src/address"
	"github.com/RiV-chain/RiV-mesh/src/config"
	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/RiV-chain/RiV-mesh/src/defaults"
	"github.com/RiV-chain/RiV-mesh/src/ipv6rwc"
	"github.com/RiV-chain/RiV-mesh/src/multicast"
	"github.com/RiV-chain/RiV-mesh/src/version"

	_ "golang.org/x/mobile/bind"
)

// Mesh mobile package is meant to "plug the gap" for mobile support, as
// Gomobile will not create headers for Swift/Obj-C etc if they have complex
// (non-native) types. Therefore for iOS we will expose some nice simple
// functions. Note that in the case of iOS we handle reading/writing to/from TUN
// in Swift therefore we use the "dummy" TUN interface instead.
type Mesh struct {
	core      core.Core
	iprwc     *ipv6rwc.ReadWriteCloser
	config    *config.NodeConfig
	multicast multicast.Multicast
	log       MobileLogger
}

// StartAutoconfigure starts a node with a randomly generated config
func (m *Mesh) StartAutoconfigure() error {
	return m.StartJSON([]byte("{}"))
}

// StartJSON starts a node with the given JSON config. You can get JSON config
// (rather than HJSON) by using the GenerateConfigJSON() function
func (m *Mesh) StartJSON(configjson []byte) error {
	logger := log.New(m.log, "", 0)
	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")
	m.config = defaults.GenerateConfig()
	if err := json.Unmarshal(configjson, &m.config); err != nil {
		return err
	}
	m.config.IfName = "none"
	if err := m.core.Start(m.config, logger); err != nil {
		logger.Errorln("An error occured starting Mesh:", err)
		return err
	}
	mtu := m.config.IfMTU
	m.iprwc = ipv6rwc.NewReadWriteCloser(&m.core)
	if m.iprwc.MaxMTU() < mtu {
		mtu = m.iprwc.MaxMTU()
	}
	m.iprwc.SetMTU(mtu)
	if len(m.config.MulticastInterfaces) > 0 {
		if err := m.multicast.Init(&m.core, m.config, logger, nil); err != nil {
			logger.Errorln("An error occurred initialising multicast:", err)
			return err
		}
		if err := m.multicast.Start(); err != nil {
			logger.Errorln("An error occurred starting multicast:", err)
			return err
		}
	}
	return nil
}

// Send sends a packet to Mesh. It should be a fully formed
// IPv6 packet
func (m *Mesh) Send(p []byte) error {
	if m.iprwc == nil {
		return nil
	}
	_, _ = m.iprwc.Write(p)
	return nil
}

// Recv waits for and reads a packet coming from Mesh. It
// will be a fully formed IPv6 packet
func (m *Mesh) Recv() ([]byte, error) {
	if m.iprwc == nil {
		return nil, nil
	}
	var buf [65535]byte
	n, _ := m.iprwc.Read(buf[:])
	return buf[:n], nil
}

// Stop the mobile Mesh instance
func (m *Mesh) Stop() error {
	logger := log.New(m.log, "", 0)
	logger.EnableLevel("info")
	logger.Infof("Stop the mobile Mesh instance %s", "")
	if err := m.multicast.Stop(); err != nil {
		return err
	}
	m.core.Stop()
	return nil
}

// GenerateConfigJSON generates mobile-friendly configuration in JSON format
func GenerateConfigJSON() []byte {
	nc := defaults.GenerateConfig()
	nc.IfName = "none"
	if json, err := json.Marshal(nc); err == nil {
		return json
	}
	return nil
}

// GetAddressString gets the node's IPv6 address
func (m *Mesh) GetAddressString() string {
	ip := m.core.Address()
	return ip.String()
}

// GetSubnetString gets the node's IPv6 subnet in CIDR notation
func (m *Mesh) GetSubnetString() string {
	subnet := m.core.Subnet()
	return subnet.String()
}

// GetPublicKeyString gets the node's public key in hex form
func (m *Mesh) GetPublicKeyString() string {
	return hex.EncodeToString(m.core.GetSelf().Key)
}

// GetCoordsString gets the node's coordinates
func (m *Mesh) GetCoordsString() string {
	return fmt.Sprintf("%v", m.core.GetSelf().Coords)
}

func (m *Mesh) GetPeersJSON() (result string) {
	peers := []struct {
		core.Peer
		IP string
	}{}
	for _, v := range m.core.GetPeers() {
		a := address.AddrForKey(v.Key)
		ip := net.IP(a[:]).String()
		peers = append(peers, struct {
			core.Peer
			IP string
		}{
			Peer: v,
			IP:   ip,
		})
	}
	if res, err := json.Marshal(peers); err == nil {
		return string(res)
	} else {
		return "{}"
	}
}

func (m *Mesh) GetDHTJSON() (result string) {
	if res, err := json.Marshal(m.core.GetDHT()); err == nil {
		return string(res)
	} else {
		return "{}"
	}
}

// GetMTU returns the configured node MTU. This must be called AFTER Start.
func (m *Mesh) GetMTU() int {
	return int(m.core.MTU())
}

func GetVersion() string {
	return version.BuildVersion()
}
