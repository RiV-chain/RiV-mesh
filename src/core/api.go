package core

import (
	"crypto/ed25519"
	"sync/atomic"
	"time"

	"encoding/json"

	"net"
	"net/url"

	"github.com/Arceliar/ironwood/network"
	"github.com/Arceliar/ironwood/types"
	"github.com/Arceliar/phony"
)

type LabelInfo struct {
	Sig    []byte
	Domain types.Domain
	Root   types.Domain
	Seq    uint64
	Beacon uint64
	Path   []uint64
}

type SelfInfo struct {
	Domain         types.Domain
	PrivateKey     ed25519.PrivateKey
	Tld            string
	RoutingEntries uint64
}

type PeerInfo struct {
	Domain        types.Domain
	Root          types.Domain
	URI           string
	Up            bool
	Inbound       bool
	LastError     error
	LastErrorTime time.Time
	Port          uint64
	Priority      uint8
	RXBytes       uint64
	TXBytes       uint64
	Uptime        time.Duration
	RemoteIp      string
}

type TreeEntryInfo struct {
	IPAddress string
	Domain    string
	Parent    string
	Sequence  uint64
}

type PathEntryInfo struct {
	IPAddress string
	Domain    string
	Path      []uint64
	Sequence  uint64
}

type SessionInfo struct {
	IPAddress string
	Domain    string
	RXBytes   uint64
	TXBytes   uint64
	Uptime    time.Duration
}

func (c *Core) GetLabel() LabelInfo {
	var label LabelInfo
	//s := c.PacketConn.PacketConn.Debug.GetLabel()
	//label.Domain = s.Domain
	//label.Root = s.Root
	//label.Seq = s.Seq
	//label.Beacon = s.Beacon
	//label.Path = s.Path
	//label.Sig = s.Sig
	return label
}

func (c *Core) GetSelf() SelfInfo {
	var self SelfInfo
	s := c.PacketConn.PacketConn.Debug.GetSelf()
	self.Domain = s.Domain
	self.PrivateKey = c.secret
	self.RoutingEntries = s.RoutingEntries
	self.Tld = c.GetDdnsServer().Tld
	return self
}

func (c *Core) GetPeers() []PeerInfo {
	peers := []PeerInfo{}
	conns := map[net.Conn]network.DebugPeerInfo{}
	iwpeers := c.PacketConn.PacketConn.Debug.GetPeers()
	for _, p := range iwpeers {
		conns[p.Conn] = p
	}

	phony.Block(&c.links, func() {
		for info, state := range c.links._links {
			var peerinfo PeerInfo
			var conn net.Conn
			peerinfo.URI = info.uri
			peerinfo.LastError = state._err
			peerinfo.LastErrorTime = state._errtime
			if c := state._conn; c != nil {
				conn = c
				peerinfo.Up = true
				peerinfo.Inbound = state.linkType == linkTypeIncoming
				peerinfo.RXBytes = atomic.LoadUint64(&c.rx)
				peerinfo.TXBytes = atomic.LoadUint64(&c.tx)
				peerinfo.Uptime = time.Since(c.up)
			}
			if p, ok := conns[conn]; ok {
				peerinfo.Domain = p.Domain
				peerinfo.RemoteIp = p.Conn.RemoteAddr().String()
				peerinfo.Root = p.Root
				peerinfo.Port = p.Port
				peerinfo.Priority = p.Priority
				peers = append(peers, peerinfo)
			}
		}
	})

	return peers
}

func (c *Core) GetTree() []TreeEntryInfo {
	var trees []TreeEntryInfo
	ts := c.PacketConn.PacketConn.Debug.GetTree()
	for _, t := range ts {
		addr := c.AddrForDomain(t.Domain)
		var info TreeEntryInfo
		info.IPAddress = net.IP(addr[:]).String()
		info.Domain = string(t.Domain.GetNormalizedName())
		info.Parent = string(t.Parent.GetNormalizedName())
		info.Sequence = t.Sequence
		trees = append(trees, info)
	}
	return trees
}

func (c *Core) GetPaths() []PathEntryInfo {
	var paths []PathEntryInfo
	ps := c.PacketConn.PacketConn.Debug.GetPaths()
	for _, p := range ps {
		addr := c.AddrForDomain(p.Domain)
		var info PathEntryInfo
		info.IPAddress = net.IP(addr[:]).String()
		info.Domain = string(p.Domain.GetNormalizedName())
		info.Sequence = p.Sequence
		info.Path = p.Path
		paths = append(paths, info)
	}
	return paths
}

func (c *Core) GetSessions() []SessionInfo {
	var sessions []SessionInfo
	ss := c.PacketConn.Debug.GetSessions()
	for _, s := range ss {
		addr := c.AddrForDomain(s.Domain)
		var info SessionInfo
		info.IPAddress = net.IP(addr[:]).String()
		info.Domain = string(s.Domain.GetNormalizedName())
		info.RXBytes = s.RX
		info.TXBytes = s.TX
		info.Uptime = s.Uptime
		sessions = append(sessions, info)
	}
	return sessions
}

// Listen starts a new listener (either TCP or TLS). The input should be a url.URL
// parsed from a string of the form e.g. "tcp://a.b.c.d:e". In the case of a
// link-local address, the interface should be provided as the second argument.
func (c *Core) Listen(u *url.URL, sintf string) (*Listener, error) {
	return c.links.listen(u, sintf)
}

// Address gets the IPv6 address of the Mesh node. This is always a /128
// address. The IPv6 address is only relevant when the node is operating as an
// IP router and often is meaningless when embedded into an application, unless
// that application also implements either VPN functionality or deals with IP
// packets specifically.
func (c *Core) Address() net.IP {
	addr := net.IP(c.AddrForDomain(c.public)[:])
	return addr
}

// Subnet gets the routed IPv6 subnet of the Mesh node. This is always a
// /64 subnet. The IPv6 subnet is only relevant when the node is operating as an
// IP router and often is meaningless when embedded into an application, unless
// that application also implements either VPN functionality or deals with IP
// packets specifically.
func (c *Core) Subnet() net.IPNet {
	subnet := c.SubnetForDomain(c.public)[:]
	subnet = append(subnet, 0, 0, 0, 0, 0, 0, 0, 0)
	return net.IPNet{IP: subnet, Mask: net.CIDRMask(64, 128)}
}

// SetLogger sets the output logger of the Mesh node after startup. This
// may be useful if you want to redirect the output later. Note that this
// expects a Logger from the github.com/gologme/log package and not from Go's
// built-in log package.
func (c *Core) SetLogger(log Logger) {
	c.log = log
}

// AddPeer adds a peer. This should be specified in the peer URI format, e.g.:
//
//	tcp://a.b.c.d:e
//	socks://a.b.c.d:e/f.g.h.i:j
//
// This adds the peer to the peer list, so that they will be called again if the
// connection drops.
func (c *Core) AddPeer(u *url.URL, sintf string) error {
	return c.links.add(u, sintf, linkTypePersistent)
}

// RemovePeer removes a peer. The peer should be specified in URI format, see AddPeer.
// The peer is not disconnected immediately.
func (c *Core) RemovePeer(url *url.URL, sourceInterface string) error {
	return c.links.remove(url, sourceInterface, linkTypePersistent)
}

// RemovePeers removes all peers.
// The peers are not disconnected immediately.
func (c *Core) RemovePeers() error {
	return c.links.removeAll()
}

// CallPeer calls a peer once. This should be specified in the peer URI format,
// e.g.:
//
//	tcp://a.b.c.d:e
//	socks://a.b.c.d:e/f.g.h.i:j
//
// This does not add the peer to the peer list, so if the connection drops, the
// peer will not be called again automatically.
func (c *Core) CallPeer(u *url.URL, sintf string) error {
	return c.links.add(u, sintf, linkTypeEphemeral)
}

func (c *Core) PublicKey() ed25519.PublicKey {
	return c.public.Key.ToEd()
}

// Hack to get the admin stuff working, TODO something cleaner

type AddHandler interface {
	AddHandler(name, desc string, args []string, handlerfunc AddHandlerFunc) error
}

type AddHandlerFunc func(json.RawMessage) (interface{}, error)

// SetAdmin must be called after Init and before Start.
// It sets the admin handler for NodeInfo and the Debug admin functions.
func (c *Core) SetAdmin(a AddHandler) error {
	if err := a.AddHandler(
		"getNodeInfo", "Request nodeinfo from a remote node by its public key", []string{"key"},
		c.proto.nodeinfo.nodeInfoAdminHandler,
	); err != nil {
		return err
	}
	if err := a.AddHandler(
		"debug_remoteGetSelf", "Debug use only", []string{"key"},
		c.proto.getSelfHandler,
	); err != nil {
		return err
	}
	if err := a.AddHandler(
		"debug_remoteGetPeers", "Debug use only", []string{"key"},
		c.proto.getPeersHandler,
	); err != nil {
		return err
	}
	if err := a.AddHandler(
		"debug_remoteGetDHT", "Debug use only", []string{"key"},
		c.proto.getDHTHandler,
	); err != nil {
		return err
	}
	return nil
}

func applyAdminCall(handlerfunc AddHandlerFunc, key string) (result map[string]any, err error) {
	var in []byte
	if in, err = json.Marshal(map[string]any{"key": key}); err != nil {
		return
	}
	var out1 any
	if out1, err = handlerfunc(in); err != nil {
		return
	}
	var out2 []byte
	if out2, err = json.Marshal(out1); err != nil {
		return
	}
	err = json.Unmarshal(out2, &result)
	return
}

func (c *Core) GetNodeInfo(key string) (result map[string]any, err error) {
	return applyAdminCall(c.proto.nodeinfo.nodeInfoAdminHandler, key)
}

func (c *Core) RemoteGetSelf(key string) (map[string]any, error) {
	return applyAdminCall(c.proto.getSelfHandler, key)
}

func (c *Core) RemoteGetPeers(key string) (map[string]any, error) {
	return applyAdminCall(c.proto.getPeersHandler, key)
}

func (c *Core) RemoteGetDHT(key string) (map[string]any, error) {
	return applyAdminCall(c.proto.getDHTHandler, key)
}
