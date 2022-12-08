package admin

import (
	"encoding/hex"

	"github.com/RiV-chain/RiV-mesh/src/version"
)

type GetSelfRequest struct{}

type GetSelfResponse struct {
	BuildName    string   `json:"build_name"`
	BuildVersion string   `json:"build_version"`
	PublicKey    string   `json:"key"`
	PrivateKey   string   `json:"private_key"`
	IPAddress    string   `json:"address"`
	Coords       []uint64 `json:"coords"`
	Subnet       string   `json:"subnet"`
}

func (a *AdminSocket) getSelfHandler(req *GetSelfRequest, res *GetSelfResponse) error {
	self := a.core.GetSelf()
	snet := a.core.Subnet()
	res.BuildName = version.BuildName()
	res.BuildVersion = version.BuildVersion()
	res.PublicKey = hex.EncodeToString(self.Key[:])
	res.PrivateKey = hex.EncodeToString(self.PrivateKey[:])
	res.IPAddress = a.core.Address().String()
	res.Subnet = snet.String()
	res.Coords = self.Coords
	return nil
}
