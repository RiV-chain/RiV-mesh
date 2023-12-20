package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Arceliar/ironwood/types"
	iwt "github.com/Arceliar/ironwood/types"
	"github.com/Arceliar/phony"
	//"github.com/RiV-chain/RiV-mesh/src/address"
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

type keyArray types.PublicKey

type protoHandler struct {
	phony.Inbox

	core     *Core
	nodeinfo nodeinfo

	selfRequests  map[types.PublicKey]*reqInfo
	peersRequests map[types.PublicKey]*reqInfo
	dhtRequests   map[types.PublicKey]*reqInfo
}

func (p *protoHandler) init(core *Core) {
	p.core = core
	p.nodeinfo.init(p)

	p.selfRequests = make(map[types.PublicKey]*reqInfo)
	p.peersRequests = make(map[types.PublicKey]*reqInfo)
	p.dhtRequests = make(map[types.PublicKey]*reqInfo)
}

// Common functions

func (p *protoHandler) handleProto(from phony.Actor, key iwt.Domain, bs []byte) {
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

func (p *protoHandler) handleDebug(from phony.Actor, key iwt.Domain, bs []byte) {
	p.Act(from, func() {
		p._handleDebug(key, bs)
	})
}

func (p *protoHandler) _handleDebug(domain iwt.Domain, bs []byte) {
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

func (p *protoHandler) _sendDebug(key iwt.Domain, dType uint8, data []byte) {
	bs := append([]byte{typeSessionProto, typeProtoDebug, dType}, data...)
	_, _ = p.core.PacketConn.WriteTo(bs, iwt.Addr(key))
}

// Get self

func (p *protoHandler) sendGetSelfRequest(domain iwt.Domain, callback func([]byte)) {
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

func (p *protoHandler) _handleGetSelfRequest(key iwt.Domain) {
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

func (p *protoHandler) _handleGetSelfResponse(domain iwt.Domain, bs []byte) {

	key := domain.Key
	if info := p.selfRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.selfRequests, key)
	}
}

// Get peers

func (p *protoHandler) sendGetPeersRequest(domain iwt.Domain, callback func([]byte)) {
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

func (p *protoHandler) _handleGetPeersRequest(domain iwt.Domain) {
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

func (p *protoHandler) _handleGetPeersResponse(domain iwt.Domain, bs []byte) {
	key := domain.Key
	if info := p.peersRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.peersRequests, key)
	}
}

// Get DHT

func (p *protoHandler) sendGetDHTRequest(domain iwt.Domain, callback func([]byte)) {
	p.Act(nil, func() {
		key := domain.Key
		if info := p.dhtRequests[key]; info != nil {
			info.timer.Stop()
			delete(p.dhtRequests, key)
		}
		info := new(reqInfo)
		info.callback = callback
		info.timer = time.AfterFunc(time.Minute, func() {
			p.Act(nil, func() {
				if p.dhtRequests[key] == info {
					delete(p.dhtRequests, key)
				}
			})
		})
		p.dhtRequests[key] = info
		p._sendDebug(domain, typeDebugGetTreeRequest, nil)
	})
}

func (p *protoHandler) _handleGetTreeRequest(domain iwt.Domain) {
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

func (p *protoHandler) _handleGetTreeResponse(domain iwt.Domain, bs []byte) {
	key := domain.Key
	if info := p.dhtRequests[key]; info != nil {
		info.timer.Stop()
		info.callback(bs)
		delete(p.dhtRequests, key)
	}
}

// Admin socket stuff for "Get self"

type DebugGetSelfRequest struct {
	Key iwt.Domain `json:"key"`
}

type DebugGetSelfResponse map[string]interface{}

func (p *protoHandler) getSelfHandler(in json.RawMessage) (interface{}, error) {
	var req DebugGetSelfRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}
	ch := make(chan []byte, 1)
	p.sendGetSelfRequest(req.Key, func(info []byte) {
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
		ip := net.IP(p.core.AddrForDomain(req.Key)[:])
		res := DebugGetSelfResponse{ip.String(): msg}
		return res, nil
	}
}

// Admin socket stuff for "Get peers"

type DebugGetPeersRequest struct {
	Key iwt.Domain `json:"key"`
}

type DebugGetPeersResponse map[string]interface{}

func (p *protoHandler) getPeersHandler(in json.RawMessage) (interface{}, error) {
	var req DebugGetPeersRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}
	key := req.Key.Key
	ch := make(chan []byte, 1)
	p.sendGetPeersRequest(req.Key, func(info []byte) {
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
		ip := net.IP(p.core.AddrForDomain(req.Key)[:])
		res := DebugGetPeersResponse{ip.String(): msg}
		return res, nil
	}
}

// Admin socket stuff for "Get DHT"

type DebugGetDHTRequest struct {
	Key iwt.Domain `json:"key"`
}

type DebugGetDHTResponse map[string]interface{}

func (p *protoHandler) getDHTHandler(in json.RawMessage) (interface{}, error) {
	var req DebugGetDHTRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}
	key := req.Key.Key
	ch := make(chan []byte, 1)
	p.sendGetDHTRequest(req.Key, func(info []byte) {
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
		ip := net.IP(p.core.AddrForDomain(req.Key)[:])
		res := DebugGetDHTResponse{ip.String(): msg}
		return res, nil
	}
}
