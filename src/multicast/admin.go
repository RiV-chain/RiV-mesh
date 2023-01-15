package multicast

type GetMulticastInterfacesRequest struct{}
type GetMulticastInterfacesResponse struct {
	Interfaces []string `json:"multicast_interfaces"`
}

func (m *Multicast) getMulticastInterfacesHandler(req *GetMulticastInterfacesRequest, res *GetMulticastInterfacesResponse) error {
	res.Interfaces = []string{}
	for _, v := range m.Interfaces() {
		res.Interfaces = append(res.Interfaces, v.Name)
	}
	return nil
}
