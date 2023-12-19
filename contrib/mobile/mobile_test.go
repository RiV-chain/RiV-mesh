package mobile

import "testing"

func TestStartMesh(t *testing.T) {
	mesh := &Mesh{}
	if err := mesh.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start RiV-mesh: %s", err)
	}
	t.Log("Address:", mesh.GetAddressString())
	t.Log("Subnet:", mesh.GetSubnetString())
	t.Log("Routing entries:", mesh.GetRoutingEntries())
	if err := mesh.Stop(); err != nil {
		t.Fatalf("Failed to stop RiV-mesh: %s", err)
	}
}
