package tun

type GetTUNRequest struct{}
type GetTUNResponse struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name,omitempty"`
	MTU     uint64 `json:"mtu,omitempty"`
}

type TUNEntry struct {
	MTU uint64 `json:"mtu"`
}

func (t *TunAdapter) getTUNHandler(req *GetTUNRequest, res *GetTUNResponse) error {
	res.Enabled = t.isEnabled
	if !t.isEnabled {
		return nil
	}
	res.Name = t.Name()
	res.MTU = t.MTU()
	return nil
}
