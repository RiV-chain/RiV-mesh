package core

import (
	"crypto/ed25519"
	"fmt"
	"net/url"
)

func (c *Core) _applyOption(opt SetupOption) (err error) {
	switch v := opt.(type) {
	case Domain:
		c.config.domain = v
	case Peer:
		u, err := url.Parse(v.URI)
		if err != nil {
			return fmt.Errorf("unable to parse peering URI: %w", err)
		}
		err = c.links.add(u, v.SourceInterface, linkTypePersistent)
		switch err {
		case ErrLinkAlreadyConfigured:
			// Don't return this error, otherwise we'll panic at startup
			// if there are multiple of the same peer configured
			return nil
		default:
			return err
		}
	case ListenAddress:
		c.config._listeners[v] = struct{}{}
	case NodeInfo:
		c.config.nodeinfo = v
	case NodeInfoPrivacy:
		c.config.nodeinfoPrivacy = v
	case NetworkDomain:
		c.config.networkdomain = v
	case AllowedPublicKey:
		pk := [32]byte{}
		copy(pk[:], v)
		c.config._allowedPublicKeys[pk] = struct{}{}
	case DDnsServer:
		c.config.ddnsserver = v
	}
	return
}

type SetupOption interface {
	isSetupOption()
}

type Domain string

type ListenAddress string
type Peer struct {
	URI             string
	SourceInterface string
}
type NodeInfo map[string]interface{}
type NodeInfoPrivacy bool
type NetworkDomain struct {
	Prefix string
}
type AllowedPublicKey ed25519.PublicKey

type DDnsServer struct {
	Tld             string
	ListenAddress   string
	UpstreamServers []string
}

func (a Domain) isSetupOption()           {}
func (a ListenAddress) isSetupOption()    {}
func (a Peer) isSetupOption()             {}
func (a NodeInfo) isSetupOption()         {}
func (a NodeInfoPrivacy) isSetupOption()  {}
func (a NetworkDomain) isSetupOption()    {}
func (a AllowedPublicKey) isSetupOption() {}
func (a DDnsServer) isSetupOption()       {}
