package core

import "errors"

// In-band packet types
const (
	typeSessionDummy = iota // nolint:deadcode,varcheck
	typeSessionTraffic
	typeSessionProto
)

// Protocol packet types
const (
	typeProtoDummy = iota
	typeProtoNodeInfoRequest
	typeProtoNodeInfoResponse
	typeProtoDebug = 255
)

var ErrTimeout = errors.New("Operation timeout")
