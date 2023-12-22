package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"time"

	iwt "github.com/Arceliar/ironwood/types"
	"github.com/Arceliar/phony"
)

const (
	typeDebugDummy = iota
	typeDebugGetSelfRequest
	typeDebugGetSelfResponse
	typeDebugGetPeersRequest
	typeDebugGetPeersResponse
	typeDebugGetTreeRequest
	typeDebugGetTreeResponse
)

type reqInfo struct {
	callback func([]byte)
	timer    *time.Timer // time.AfterFunc cleanup
}

type protoHandler struct {
	phony.Inbox

	core     *Core
	nodeinfo nodeinfo

	selfRequests  map[iwt.PublicKey]*reqInfo
	peersRequests map[iwt.PublicKey]*reqInfo
	treeRequests  map[iwt.PublicKey]*reqInfo
}

func (p *protoHandler) init(core *Core) {
	p.core = core
	p.nodeinfo.init(p)

	p.selfRequests = make(map[iwt.PublicKey]*reqInfo)
	p.peersRequests = make(map[iwt.PublicKey]*reqInfo)
}

// Common functions

func (p *protoHandler) handleProto(from phony.Actor, key iwt.Addr, bs []byte) {
	if len(bs) == 0 {
		return
	}
	switch bs[0] {
	case typeProtoDummy:
	case typeProtoNodeInfoRequest:
		p.nodeinfo.handleReq(p, key)
	case typeProtoNodeInfoResponse:
		p.nodeinfo.handleRes(p, key, bs[1:])
	case typeProtoDebug:
		p.handleDebug(from, key, bs[1:])
	}
}

func (p *protoHandler) handleDebug(from phony.Actor, key iwt.Addr, bs []byte) {
	p.Act(from, func() {
		p._handleDebug(key, bs)
	})
}

func (p *protoHandler) _handleDebug(domain iwt.Addr, bs []byte) {
	if len(bs) == 0 {
		return
	}

	switch bs[0] {
	case typeDebugDummy:
	case typeDebugGetSelfRequest:
		p._handleGetSelfRequest(domain)
	case typeDebugGetSelfResponse:
		p._handleGetSelfResponse(domain, bs[1:])
	case typeDebugGetPeersRequest:
		p._handleGetPeersRequest(domain)
	case typeDebugGetPeersResponse:
		p._handleGetPeersResponse(domain, bs[1:])
	case typeDebugGetTreeRequest:
		p._handleGetTreeRequest(domain)
	case typeDebugGetTreeResponse:
		p._handleGetTreeResponse(domain, bs[1:])
	}
}

func (p *protoHandler) _sendDebug(key iwt.Addr, dType uint8, data []byte) {
	bs := append([]byte{typeSessionProto, typeProtoDebug, dType}, data...)
	_, _ = p.core.PacketConn.WriteTo(bs, key)
}

// Get self

func (p *protoHandler) sendGetSelfRequest(domain iwt.Addr, callback func([]byte)) {
	p.Act(nil, func() {
		key := domain.Key
		if info := p.selfRequests[key]; info != nil {
			info.timer.Stop()
			delete(p.selfRequests, key)
		}
		info := new(reqInfo)
		info.callback = callback
		info.timer = time.AfterFunc(time.Minute, func() {
			p.Act(nil, func() {
				if p.selfRequests[key] == info {
					delete(p.selfRequests, key)
				}
			})
		})
		p.selfRequests[key] = info
		p._sendDebug(domain, typeDebugGetSelfRequest, nil)
	})
}

func (p *protoHandler) _handleGetSelfRequest(key iwt.Addr) {
	self := p.core.GetSelf()
	res := map[string]string{
		"key":    hex.EncodeToString(self.Domain.Key.ToSlice()),
		"domain": string(self.Domain.GetNormalizedName()),
		"tld":    self.Tld,
		"coords": fmt.Sprintf("%v", self.RoutingEntries),
	}
	bs, err := json.Marshal(res) // FIXME this puts keys in base64, not hex
	if err != nil {
		return
	}
	p._sendDebug(key, typeDebugGetSelfResponse, bs)
}

func (p *protoHandler) _handleGetSelfResponse(domain iwt.Addr, bs []byte) {

	key := domain.Key
	if info := p.selfRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.selfRequests, key)
	}
}

// Get peers

func (p *protoHandler) sendGetPeersRequest(domain iwt.Addr, callback func([]byte)) {
	p.Act(nil, func() {
		key := domain.Key
		if info := p.peersRequests[key]; info != nil {
			info.timer.Stop()
			delete(p.peersRequests, key)
		}
		info := new(reqInfo)
		info.callback = callback
		info.timer = time.AfterFunc(time.Minute, func() {
			p.Act(nil, func() {
				if p.peersRequests[key] == info {
					delete(p.peersRequests, key)
				}
			})
		})
		p.peersRequests[key] = info
		p._sendDebug(domain, typeDebugGetPeersRequest, nil)
	})
}

func (p *protoHandler) _handleGetPeersRequest(domain iwt.Addr) {
	peers := p.core.GetPeers()
	var bs []byte
	for _, pinfo := range peers {
		tmp := append(bs, pinfo.Domain.Key[:]...)
		const responseOverhead = 2 // 1 debug type, 1 getpeers type
		if uint64(len(tmp))+responseOverhead > p.core.MTU() {
			break
		}
		bs = tmp
	}
	p._sendDebug(domain, typeDebugGetPeersResponse, bs)
}

func (p *protoHandler) _handleGetPeersResponse(domain iwt.Addr, bs []byte) {
	key := domain.Key
	if info := p.peersRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.peersRequests, key)
	}
}

func (p *protoHandler) _handleGetTreeRequest(domain iwt.Addr) {
	dinfos := p.core.GetTree()
	var bs []byte
	for _, dinfo := range dinfos {
		tmp := append(bs, dinfo.Domain[:]...)
		const responseOverhead = 2 // 1 debug type, 1 gettree type
		if uint64(len(tmp))+responseOverhead > p.core.MTU() {
			break
		}
		bs = tmp
	}
	p._sendDebug(domain, typeDebugGetTreeResponse, bs)
}

func (p *protoHandler) _handleGetTreeResponse(domain iwt.Addr, bs []byte) {
	key := domain.Key
	if info := p.treeRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.treeRequests, key)
	}
}

// Admin socket stuff for "Get self"

type DebugGetSelfRequest struct {
	Name string `json:"name"`
}

type DebugGetSelfResponse map[string]interface{}

func (p *protoHandler) getSelfHandler(in json.RawMessage) (interface{}, error) {
	var req DebugGetSelfRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}
	ch := make(chan []byte, 1)
	var key [32]byte
	domain := iwt.NewDomain(req.Name, key[:])
	p.sendGetSelfRequest(iwt.Addr(domain), func(info []byte) {
		ch <- info
	})
	timer := time.NewTimer(6 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil, ErrTimeout
	case info := <-ch:
		var msg json.RawMessage
		if err := msg.UnmarshalJSON(info); err != nil {
			return nil, err
		}
		ip := net.IP(p.core.AddrForDomain(domain)[:])
		res := DebugGetSelfResponse{ip.String(): msg}
		return res, nil
	}
}

// Admin socket stuff for "Get peers"

type DebugGetPeersRequest struct {
	Name string `json:"name"`
}

type DebugGetPeersResponse map[string]interface{}

func (p *protoHandler) getPeersHandler(in json.RawMessage) (interface{}, error) {
	var req DebugGetPeersRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}

	ch := make(chan []byte, 1)
	var key [32]byte
	domain := iwt.NewDomain(req.Name, key[:])
	p.sendGetPeersRequest(iwt.Addr(domain), func(info []byte) {
		ch <- info
	})
	timer := time.NewTimer(6 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil, ErrTimeout
	case info := <-ch:
		ks := make(map[string][]string)
		bs := info
		for len(bs) >= len(key) {
			ks["keys"] = append(ks["keys"], hex.EncodeToString(bs[:len(key)]))
			bs = bs[len(key):]
		}
		js, err := json.Marshal(ks)
		if err != nil {
			return nil, err
		}
		var msg json.RawMessage
		if err := msg.UnmarshalJSON(js); err != nil {
			return nil, err
		}
		ip := net.IP(p.core.AddrForDomain(domain)[:])
		res := DebugGetPeersResponse{ip.String(): msg}
		return res, nil
	}
}
