package types

import (
	"github.com/Arceliar/ironwood/types"
)

type DDNS interface {

	// Verifies whether the domain is registered
	IsRegistered(domain types.Domain) bool

	// Broadcast the registration to peers recursively
	Register(domain types.Domain) error

	// Sync DB
	Sync() error

	// Save to local DB
	Save(domain types.Domain) error
}
