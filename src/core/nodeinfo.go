package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/Arceliar/ironwood/types"
	iwt "github.com/Arceliar/ironwood/types"
	"github.com/Arceliar/phony"

	//"github.com/RiV-chain/RiV-mesh/src/crypto"

	"github.com/RiV-chain/RiV-mesh/src/version"
)

type nodeinfo struct {
	phony.Inbox
	proto      *protoHandler
	myNodeInfo json.RawMessage
	callbacks  map[types.PublicKey]nodeinfoCallback
}

type nodeinfoCallback struct {
	call    func(nodeinfo json.RawMessage)
	created time.Time
}

// Initialises the nodeinfo cache/callback maps, and starts a goroutine to keep
// the cache/callback maps clean of stale entries
func (m *nodeinfo) init(proto *protoHandler) {
	m.Act(nil, func() {
		m._init(proto)
	})
}

func (m *nodeinfo) _init(proto *protoHandler) {
	m.proto = proto
	m.callbacks = make(map[types.PublicKey]nodeinfoCallback)
	m._cleanup()
}

func (m *nodeinfo) _cleanup() {
	for boxPubKey, callback := range m.callbacks {
		if time.Since(callback.created) > time.Minute {
			delete(m.callbacks, boxPubKey)
		}
	}
	time.AfterFunc(time.Second*30, func() {
		m.Act(nil, m._cleanup)
	})
}

func (m *nodeinfo) _addCallback(sender types.PublicKey, call func(nodeinfo json.RawMessage)) {
	m.callbacks[sender] = nodeinfoCallback{
		created: time.Now(),
		call:    call,
	}
}

// Handles the callback, if there is one
func (m *nodeinfo) _callback(sender types.PublicKey, nodeinfo json.RawMessage) {
	if callback, ok := m.callbacks[sender]; ok {
		callback.call(nodeinfo)
		delete(m.callbacks, sender)
	}
}

func (m *nodeinfo) _getNodeInfo() json.RawMessage {
	return m.myNodeInfo
}

// Set the current node's nodeinfo
func (m *nodeinfo) setNodeInfo(given map[string]interface{}, privacy bool) (err error) {
	phony.Block(m, func() {
		err = m._setNodeInfo(given, privacy)
	})
	return
}

func (m *nodeinfo) _setNodeInfo(given map[string]interface{}, privacy bool) error {
	newnodeinfo := make(map[string]interface{}, len(given))
	for k, v := range given {
		newnodeinfo[k] = v
	}
	if !privacy {
		newnodeinfo["buildname"] = version.BuildName()
		newnodeinfo["buildversion"] = version.BuildVersion()
		newnodeinfo["buildplatform"] = runtime.GOOS
		newnodeinfo["buildarch"] = runtime.GOARCH
	}
	newjson, err := json.Marshal(newnodeinfo)
	switch {
	case err != nil:
		return fmt.Errorf("NodeInfo marshalling failed: %w", err)
	case len(newjson) > 16384:
		return fmt.Errorf("NodeInfo exceeds max length of 16384 bytes")
	default:
		m.myNodeInfo = newjson
		return nil
	}
}

func (m *nodeinfo) sendReq(from phony.Actor, key iwt.Domain, callback func(nodeinfo json.RawMessage)) {
	m.Act(from, func() {
		m._sendReq(key, callback)
	})
}

func (m *nodeinfo) _sendReq(domain iwt.Domain, callback func(nodeinfo json.RawMessage)) {
	if callback != nil {
		key := domain.Key
		m._addCallback(key, callback)
	}
	_, _ = m.proto.core.PacketConn.WriteTo([]byte{typeSessionProto, typeProtoNodeInfoRequest}, iwt.Addr(domain))
}

func (m *nodeinfo) handleReq(from phony.Actor, key iwt.Domain) {
	m.Act(from, func() {
		m._sendRes(key)
	})
}

func (m *nodeinfo) handleRes(from phony.Actor, domain iwt.Domain, info json.RawMessage) {
	m.Act(from, func() {
		key := domain.Key
		m._callback(key, info)
	})
}

func (m *nodeinfo) _sendRes(key iwt.Domain) {
	bs := append([]byte{typeSessionProto, typeProtoNodeInfoResponse}, m._getNodeInfo()...)
	_, _ = m.proto.core.PacketConn.WriteTo(bs, iwt.Addr(key))
}

// Admin socket stuff

type GetNodeInfoRequest struct {
	Key iwt.Domain `json:"key"`
}
type GetNodeInfoResponse map[string]json.RawMessage

func (m *nodeinfo) nodeInfoAdminHandler(in json.RawMessage) (interface{}, error) {
	var req GetNodeInfoRequest
	if err := json.Unmarshal(in, &req); err != nil {
		return nil, err
	}
	var zeros [32]byte
	if req.Key.Key.Equal(zeros) {
		return nil, fmt.Errorf("no remote public key supplied")
	}
	ch := make(chan []byte, 1)
	m.sendReq(nil, req.Key, func(info json.RawMessage) {
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
		key := hex.EncodeToString(req.Key.Key.ToSlice())
		res := GetNodeInfoResponse{key: msg}
		return res, nil
	}
}
