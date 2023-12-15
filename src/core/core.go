package core

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	iwe "github.com/Arceliar/ironwood/encrypted"
	iwn "github.com/Arceliar/ironwood/network"
	iwt "github.com/Arceliar/ironwood/types"

	"github.com/Arceliar/phony"
	"github.com/gologme/log"
	signals "github.com/vorot93/golang-signals"

	"github.com/RiV-chain/RiV-mesh/src/version"
)

// The Core object represents the Mesh node. You should create a Core
// object for each Mesh node you plan to run.
type Core struct {
	// This is the main data structure that holds everything else for a node
	// We're going to keep our own copy of the provided config - that way we can
	// guarantee that it will be covered by the mutex
	phony.Inbox
	*iwe.PacketConn
	ctx                context.Context
	cancel             context.CancelFunc
	secret             ed25519.PrivateKey
	public             iwt.Domain
	links              links
	proto              protoHandler
	log                Logger
	addPeerTimer       *time.Timer
	PeersChangedSignal signals.Signal
	config             struct {
		domain             Domain
		_peers             map[Peer]*linkInfo         // configurable after startup
		_listeners         map[ListenAddress]struct{} // configurable after startup
		nodeinfo           NodeInfo                   // configurable after startup
		nodeinfoPrivacy    NodeInfoPrivacy            // immutable after startup
		_allowedPublicKeys map[[32]byte]struct{}      // configurable after startup
		networkdomain      NetworkDomain              // immutable after startup
		ddnsserver         DDnsServer                 // ddns config
	}
	pathNotify func(iwt.Domain)
}

func New(secret ed25519.PrivateKey, logger Logger, opts ...SetupOption) (*Core, error) {
	c := &Core{
		log: logger,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	if c.log == nil {
		c.log = log.New(io.Discard, "", 0)
	}
	if name := version.BuildName(); name != "unknown" {
		c.log.Infoln("Build name:", name)
	}
	if version := version.BuildVersion(); version != "unknown" {
		c.log.Infoln("Build version:", version)
	}

	// Take a copy of the private key so that it is in our own memory space.
	if len(secret) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("private key is incorrect length")
	}
	c.secret = make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(c.secret, secret)
	c.config._peers = map[Peer]*linkInfo{}
	c.config._listeners = map[ListenAddress]struct{}{}
	c.config._allowedPublicKeys = map[[32]byte]struct{}{}
	for _, opt := range opts {
		c._applyOption(opt)
	}
	c.public = iwt.NewDomain(string(c.config.domain), secret.Public().(ed25519.PublicKey))
	var err error
	keyXform := func(key iwt.Domain) iwt.Domain {
		return key
	}
	if c.PacketConn, err = iwe.NewPacketConn(
		c.secret,
		c.public,
		iwn.WithBloomTransform(keyXform),
		iwn.WithPeerMaxMessageSize(65535*2),
		iwn.WithPathNotify(c.doPathNotify),
	); err != nil {
		return nil, fmt.Errorf("error creating encryption: %w", err)
	}
	if c.log == nil {
		c.log = log.New(io.Discard, "", 0)
	}
	c.proto.init(c)
	if err := c.links.init(c); err != nil {
		return nil, fmt.Errorf("error initialising links: %w", err)
	}
	if err := c.proto.nodeinfo.setNodeInfo(c.config.nodeinfo, bool(c.config.nodeinfoPrivacy)); err != nil {
		return nil, fmt.Errorf("error setting node info: %w", err)
	}
	for listenaddr := range c.config._listeners {
		u, err := url.Parse(string(listenaddr))
		if err != nil {
			c.log.Errorf("Invalid listener URI %q specified, ignoring\n", listenaddr)
			continue
		}
		if _, err = c.links.listen(u, ""); err != nil {
			c.log.Errorf("Failed to start listener %q: %s\n", listenaddr, err)
		}
	}
	c.Act(nil, c._addPeerLoop)
	return c, nil
}

func (c *Core) SetThisNodeInfo(nodeinfo NodeInfo) error {
	if err := c.proto.nodeinfo.setNodeInfo(nodeinfo, bool(c.config.nodeinfoPrivacy)); err != nil {
		return fmt.Errorf("error setting node info: %w", err)
	}
	return nil
}

func (c *Core) GetDdnsServer() DDnsServer {
	return c.config.ddnsserver
}

func (c *Core) GetThisNodeInfo() json.RawMessage {
	return c.proto.nodeinfo._getNodeInfo()
}

// If any static peers were provided in the configuration above then we should
// configure them. The loop ensures that disconnected peers will eventually
// be reconnected with.
func (c *Core) _addPeerLoop() {
	select {
	case <-c.ctx.Done():
		return
	default:
	}
	// Add peers from the Peers section
	for peer := range c.config._peers {
		go func(peer string, intf string) {
			u, err := url.Parse(peer)
			if err != nil {
				c.log.Errorln("Failed to parse peer url:", peer, err)
			}
			if err := c.CallPeer(u, intf); err != nil {
				c.log.Errorln("Failed to add peer:", err)
			}
		}(peer.URI, peer.SourceInterface) // TODO: this should be acted and not in a goroutine?
	}

	c.addPeerTimer = time.AfterFunc(time.Minute, func() {
		c.Act(nil, c._addPeerLoop)
	})
}

func (c *Core) RetryPeersNow() {
	if c.addPeerTimer != nil && !c.addPeerTimer.Stop() {
		<-c.addPeerTimer.C
	}
	c.Act(nil, c._addPeerLoop)
}

// Stop shuts down the Mesh node.
func (c *Core) Stop() {
	phony.Block(c, func() {
		c.log.Infoln("Stopping...")
		_ = c._close()
		c.log.Infoln("Stopped")
	})
}

// This function is unsafe and should only be ran by the core actor.
func (c *Core) _close() error {
	c.cancel()
	c.links.shutdown()
	err := c.PacketConn.Close()
	if c.addPeerTimer != nil {
		c.addPeerTimer.Stop()
		c.addPeerTimer = nil
	}
	return err
}

func (c *Core) MTU() uint64 {
	const sessionTypeOverhead = 1
	MTU := c.PacketConn.MTU() - sessionTypeOverhead
	if MTU > 65535 {
		MTU = 65535
	}
	return MTU
}

func (c *Core) ReadFrom(p []byte) (n int, from net.Addr, err error) {
	buf := allocBytes(int(c.PacketConn.MTU()))
	defer freeBytes(buf)
	for {
		bs := buf
		n, from, err = c.PacketConn.ReadFrom(bs)
		if err != nil {
			return 0, from, err
		}
		if n == 0 {
			continue
		}
		switch bs[0] {
		case typeSessionTraffic:
			// This is what we want to handle here
		case typeSessionProto:
			key := iwt.Domain(from.(iwt.Addr))
			data := append([]byte(nil), bs[1:n]...)
			c.proto.handleProto(nil, key, data)
			continue
		default:
			continue
		}
		bs = bs[1:n]
		copy(p, bs)
		if len(p) < len(bs) {
			n = len(p)
		} else {
			n = len(bs)
		}
		return
	}
}

func (c *Core) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	buf := allocBytes(0)
	defer freeBytes(buf)
	buf = append(buf, typeSessionTraffic)
	buf = append(buf, p...)
	n, err = c.PacketConn.WriteTo(buf, addr)
	if n > 0 {
		n -= 1
	}
	return
}

func (c *Core) doPathNotify(key iwt.Domain) {
	c.Act(nil, func() {
		if c.pathNotify != nil {
			c.pathNotify(key)
		}
	})
}

func (c *Core) SetPathNotify(notify func(iwt.Domain)) {
	c.Act(nil, func() {
		c.pathNotify = notify
	})
}

type Logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})
	Traceln(...interface{})
}
