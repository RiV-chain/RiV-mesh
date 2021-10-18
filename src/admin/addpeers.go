package admin

import (

	"fmt"
	
)

type AddPeersRequest struct {
	Peers map[string]PeerInfo `json:"peers"`
}

type AddPeersResponse struct {
	List map[string]ListEntry `json:"list"`
}

type PeerInfo struct {
	Intf string   `json:"intf"`
	Uri  string   `json:"uri"`
}

func (a *AdminSocket) addPeersHandler(req *AddPeersRequest, res *AddPeersResponse) error {

	for _, p := range req.Peers {
		// Set sane defaults
		err:=a.core.AddPeer(p.Uri, p.Intf)
		if err != nil {
			return err
		} else {
			fmt.Println("added peer %s", p.Uri)
		}
	}
	return nil
}
