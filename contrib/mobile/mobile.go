package mobile

import (
	"encoding/hex"
	"encoding/json"
	"net"
	"regexp"

	"github.com/gologme/log"

	//"github.com/RiV-chain/RiV-mesh/src/address"
	"github.com/RiV-chain/RiV-mesh/src/config"
	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/RiV-chain/RiV-mesh/src/ipv6rwc"
	"github.com/RiV-chain/RiV-mesh/src/multicast"
	"github.com/RiV-chain/RiV-mesh/src/restapi"
	"github.com/RiV-chain/RiV-mesh/src/tun"
	"github.com/RiV-chain/RiV-mesh/src/version"
)

// RiV-mesh mobile package is meant to "plug the gap" for mobile support, as
// Gomobile will not create headers for Swift/Obj-C etc if they have complex
// (non-native) types. Therefore for iOS we will expose some nice simple
// functions. Note that in the case of iOS we handle reading/writing to/from TUN
// in Swift therefore we use the "dummy" TUN interface instead.

type Mesh struct {
	core        *core.Core
	iprwc       *ipv6rwc.ReadWriteCloser
	config      *config.NodeConfig
	multicast   *multicast.Multicast
	rest_server *restapi.RestServer
	tun         *tun.TunAdapter // optional
	log         MobileLogger
	logger      *log.Logger
}

// StartAutoconfigure starts a node with a randomly generated config
func (m *Mesh) StartAutoconfigure() error {
	return m.StartJSON([]byte("{}"))
}

// StartJSON starts a node with the given JSON config. You can get JSON config
// (rather than HJSON) by using the GenerateConfigJSON() function
func (m *Mesh) StartJSON(configjson []byte) error {
	setMemLimitIfPossible()

	logger := log.New(m.log, "", 0)
	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")
	m.logger = logger
	m.config = config.GenerateConfig()
	if err := m.config.UnmarshalHJSON(configjson); err != nil {
		return err
	}
	// Setup the Mesh node itself.
	{
		options := []core.SetupOption{}
		for _, peer := range m.config.Peers {
			options = append(options, core.Peer{URI: peer})
		}
		for intf, peers := range m.config.InterfacePeers {
			for _, peer := range peers {
				options = append(options, core.Peer{URI: peer, SourceInterface: intf})
			}
		}
		for _, allowed := range m.config.AllowedPublicKeys {
			k, err := hex.DecodeString(allowed)
			if err != nil {
				panic(err)
			}
			options = append(options, core.AllowedPublicKey(k[:]))
		}
		for _, lAddr := range m.config.Listen {
			options = append(options, core.ListenAddress(lAddr))
		}
		var err error
		m.core, err = core.New(m.config.Certificate, logger, options...)
		if err != nil {
			panic(err)
		}
		address, subnet := m.core.Address(), m.core.Subnet()
		logger.Infof("Your public key is %s", hex.EncodeToString(m.core.PublicKey()))
		logger.Infof("Your IPv6 address is %s", address.String())
		logger.Infof("Your IPv6 subnet is %s", subnet.String())
	}

	// Setup the multicast module.
	if len(m.config.MulticastInterfaces) > 0 {
		var err error
		logger.Infof("Initializing multicast %s", "")
		options := []multicast.SetupOption{}
		for _, intf := range m.config.MulticastInterfaces {
			options = append(options, multicast.MulticastInterface{
				Regex:    regexp.MustCompile(intf.Regex),
				Beacon:   intf.Beacon,
				Listen:   intf.Listen,
				Port:     intf.Port,
				Priority: uint8(intf.Priority),
			})
		}
		logger.Infof("Starting multicast %s", "")
		m.multicast, err = multicast.New(m.core, m.logger, options...)
		if err != nil {
			logger.Errorln("An error occurred starting multicast:", err)
		}
	}

	// Setup the REST socket.
	{
		var err error
		if m.rest_server, err = restapi.NewRestServer(restapi.RestServerCfg{
			Core:          m.core,
			Multicast:     m.multicast,
			Log:           logger,
			ListenAddress: m.config.HttpAddress,
			WwwRoot:       m.config.WwwRoot,
			ConfigFn:      "",
		}); err != nil {
			logger.Errorln(err)
		} else {
			err = m.rest_server.Serve()
			if err != nil {
				logger.Errorln(err)
			}
		}
	}

	mtu := m.config.IfMTU
	m.iprwc = ipv6rwc.NewReadWriteCloser(m.core)
	if m.iprwc.MaxMTU() < mtu {
		mtu = m.iprwc.MaxMTU()
	}
	m.iprwc.SetMTU(mtu)
	return nil
}

// Send sends a packet to RiV-mesh. It should be a fully formed
// IPv6 packet
func (m *Mesh) Send(p []byte) error {
	if m.iprwc == nil {
		return nil
	}
	_, _ = m.iprwc.Write(p)
	return nil
}

// Send sends a packet from given buffer to RiV-mesh. From first byte up to length.
func (m *Mesh) SendBuffer(p []byte, length int) error {
	if m.iprwc == nil {
		return nil
	}
	if len(p) < length {
		return nil
	}
	_, _ = m.iprwc.Write(p[:length])
	return nil
}

// Recv waits for and reads a packet coming from RiV-mesh. It
// will be a fully formed IPv6 packet
func (m *Mesh) Recv() ([]byte, error) {
	if m.iprwc == nil {
		return nil, nil
	}
	var buf [65535]byte
	n, _ := m.iprwc.Read(buf[:])
	return buf[:n], nil
}

// Recv waits for and reads a packet coming from RiV-mesh to given buffer, returning size of packet
func (m *Mesh) RecvBuffer(buf []byte) (int, error) {
	if m.iprwc == nil {
		return 0, nil
	}
	n, _ := m.iprwc.Read(buf)
	return n, nil
}

// Stop the mobile Mesh instance
func (m *Mesh) Stop() error {
	logger := log.New(m.log, "", 0)
	logger.EnableLevel("info")
	logger.Infof("Stopping the mobile Mesh instance %s", "")
	if m.multicast != nil {
		logger.Infof("Stopping multicast %s", "")
		if err := m.multicast.Stop(); err != nil {
			return err
		}
	}
	logger.Infof("Stopping TUN device %s", "")
	if m.tun != nil {
		if err := m.tun.Stop(); err != nil {
			return err
		}
	}
	logger.Infof("Stopping Mesh core %s", "")
	m.core.Stop()
	m.rest_server.Shutdown()
	m.rest_server = nil
	return nil
}

// Retry resets the peer connection timer and tries to dial them immediately.
func (m *Mesh) RetryPeersNow() {
	m.core.RetryPeersNow()
}

// GenerateConfigJSON generates mobile-friendly configuration in JSON format
func GenerateConfigJSON() []byte {
	nc := config.GenerateConfig()
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
	return hex.EncodeToString(m.core.GetSelf().Domain.Key)
}

// GetRoutingEntries gets the number of entries in the routing table
func (m *Mesh) GetRoutingEntries() int {
	return int(m.core.GetSelf().RoutingEntries)
}

func (m *Mesh) GetPeersJSON() (result string) {
	peers := []struct {
		core.PeerInfo
		IP string
	}{}
	for _, v := range m.core.GetPeers() {
		a := m.core.AddrForDomain(v.Domain)
		ip := net.IP(a[:]).String()
		peers = append(peers, struct {
			core.PeerInfo
			IP string
		}{
			PeerInfo: v,
			IP:       ip,
		})
	}
	if res, err := json.Marshal(peers); err == nil {
		return string(res)
	} else {
		return "{}"
	}
}

func (m *Mesh) GetPathsJSON() (result string) {
	if res, err := json.Marshal(m.core.GetPaths()); err == nil {
		return string(res)
	} else {
		return "{}"
	}
}

func (m *Mesh) GetTreeJSON() (result string) {
	if res, err := json.Marshal(m.core.GetTree()); err == nil {
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
