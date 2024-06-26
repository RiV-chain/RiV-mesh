package mobile

import "testing"

func TestStartMesh(t *testing.T) {
	mesh := &Mesh{}
	if err := mesh.StartAutoconfigure(14); err != nil {
		t.Fatalf("Failed to start RiV-mesh: %s", err)
	}
	t.Log("Address:", mesh.GetAddressString())
	t.Log("Subnet:", mesh.GetSubnetString())
	t.Log("Coords:", mesh.GetCoordsString())
	if err := mesh.Stop(); err != nil {
		t.Fatalf("Failed to stop RiV-mesh: %s", err)
	}
}
