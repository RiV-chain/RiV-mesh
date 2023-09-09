package core

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/Arceliar/ironwood/types"
)

func (c *Core) TestAddress_Address_IsValid(t *testing.T) {
	var address Address
	_, err := rand.Read(address[:])
	if err != nil {
		t.Fatal(err)
	}
	address[0] = 0

	if c.IsValidAddress(address) {
		t.Fatal("invalid address marked as valid")
	}

	address[0] = 0xfd

	if c.IsValidAddress(address) {
		t.Fatal("invalid address marked as valid")
	}

	address[0] = 0xfc

	if !c.IsValidAddress(address) {
		t.Fatal("valid address marked as invalid")
	}
}

func (c *Core) TestAddress_Subnet_IsValid(t *testing.T) {
	var subnet Subnet
	_, err := rand.Read(subnet[:])
	if err != nil {
		t.Fatal(err)
	}
	subnet[0] = 0

	if c.IsValidSubnet(subnet) {
		t.Fatal("invalid subnet marked as valid")
	}

	subnet[0] = 0xfc

	if c.IsValidSubnet(subnet) {
		t.Fatal("invalid subnet marked as valid")
	}

	subnet[0] = 0xfd

	if !c.IsValidSubnet(subnet) {
		t.Fatal("valid subnet marked as invalid")
	}
}

func TestAddrForDomain(t *testing.T) {
	expectedIPv6Address := Address{
		0xfc, 0x8, 0xe6, 0x97, 0x43, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
	// Mock Core instance with GetPrefix method
	core := &Core{
		ctx: context.Background(),
		config: struct {
			domain             Domain
			_peers             map[Peer]*linkInfo
			_listeners         map[ListenAddress]struct{}
			nodeinfo           NodeInfo
			nodeinfoPrivacy    NodeInfoPrivacy
			_allowedPublicKeys map[[32]byte]struct{}
			networkdomain      NetworkDomain
			ddnsserver         DDnsServer
		}{
			networkdomain: NetworkDomain{
				Prefix: "fc",
			},
		},
	}

	name := "example"
	key := ed25519.PublicKey{
		189, 186, 207, 216, 34, 64, 222, 61, 205, 18, 57, 36, 203, 181, 82, 86,
		251, 141, 171, 8, 170, 152, 227, 5, 82, 138, 184, 79, 65, 158, 110, 251,
	}

	domain := types.NewDomain(name, key)
	addr := core.AddrForDomain(domain)

	if addr == nil {
		t.Errorf("Expected non-nil address, but got nil.")
	}

	if !bytes.Equal(addr[:], expectedIPv6Address[:]) {
		t.Errorf("Expected IPv6 address does not match encoded IPv6 address.")
	}
}

func TestGetAddressDomain(t *testing.T) {
	domainName := "example"
	var bytes [ed25519.PublicKeySize]byte
	expectedDomainName := types.NewDomain(domainName, bytes[:])
	// Mock Core instance with GetPrefix method
	core := &Core{}

	ipv6Address := Address{
		0xfc, 0x8, 0xe6, 0x97, 0x43, 0xa3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}

	domain := core.GetAddressDomain(ipv6Address)

	if domain.Equal(types.Domain{}) {
		t.Errorf("Expected non-empty domain, but got empty.")
	}

	if !expectedDomainName.Equal(domain) {
		t.Errorf("Expected domain name does not match decoded domain name.")
	}
}

func TestSubnetForDomain(t *testing.T) {
	expectedIPv6Address := Subnet{
		0xfd, 0x8, 0xe6, 0x97, 0x43, 0xa3, 0x0, 0x0,
	}
	// Mock Core instance with GetPrefix method
	core := &Core{
		ctx: context.Background(),
		config: struct {
			domain             Domain
			_peers             map[Peer]*linkInfo
			_listeners         map[ListenAddress]struct{}
			nodeinfo           NodeInfo
			nodeinfoPrivacy    NodeInfoPrivacy
			_allowedPublicKeys map[[32]byte]struct{}
			networkdomain      NetworkDomain
			ddnsserver         DDnsServer
		}{
			networkdomain: NetworkDomain{
				Prefix: "fc",
			},
		},
	}

	name := "example"
	key := ed25519.PublicKey{
		189, 186, 207, 216, 34, 64, 222, 61, 205, 18, 57, 36, 203, 181, 82, 86,
		251, 141, 171, 8, 170, 152, 227, 5, 82, 138, 184, 79, 65, 158, 110, 251,
	}

	domain := types.NewDomain(name, key)
	addr := core.SubnetForDomain(domain)

	if addr == nil {
		t.Errorf("Expected non-nil address, but got nil.")
	}

	if !bytes.Equal(addr[:], expectedIPv6Address[:]) {
		t.Errorf("Expected IPv6 subnet address does not match encoded IPv6 subnet address.")
	}
}

func TestGetSubnetDomain(t *testing.T) {
	domainName := "example"
	var bytes [ed25519.PublicKeySize]byte
	expectedDomainName := types.NewDomain(domainName, bytes[:])
	// Mock Core instance with GetPrefix method
	core := &Core{}

	ipv6Address := Subnet{
		0xfc, 0x8, 0xe6, 0x97, 0x43, 0xa3, 0x0, 0x0,
	}

	domain := core.GetSubnetDomain(ipv6Address)

	if domain.Equal(types.Domain{}) {
		t.Errorf("Expected non-empty domain, but got empty.")
	}

	if !expectedDomainName.Equal(domain) {
		t.Errorf("Expected domain name does not match decoded domain name.")
	}
}

func TestMaxLengthDomain(t *testing.T) {

	// Mock Core instance with GetPrefix method
	core := &Core{
		ctx: context.Background(),
		config: struct {
			domain             Domain
			_peers             map[Peer]*linkInfo
			_listeners         map[ListenAddress]struct{}
			nodeinfo           NodeInfo
			nodeinfoPrivacy    NodeInfoPrivacy
			_allowedPublicKeys map[[32]byte]struct{}
			networkdomain      NetworkDomain
			ddnsserver         DDnsServer
		}{
			networkdomain: NetworkDomain{
				Prefix: "fc",
			},
		},
	}

	name := "veryverylongdomainnamez"
	key := ed25519.PublicKey{
		189, 186, 207, 216, 34, 64, 222, 61, 205, 18, 57, 36, 203, 181, 82, 86,
		251, 141, 171, 8, 170, 152, 227, 5, 82, 138, 184, 79, 65, 158, 110, 251,
	}

	domain := types.NewDomain(name, key)
	addr := *core.AddrForDomain(domain)

	expectedDomain := core.GetAddressDomain(addr)

	if expectedDomain.Equal(types.Domain{}) {
		t.Errorf("Expected non-empty domain, but got empty.")
	}

	if !bytes.Equal(expectedDomain.GetNormalizedName(), []byte(name)) {
		t.Errorf("Expected domain name does not match decoded domain name.")
	}

}

func TestIsValidDomain(t *testing.T) {
	testCases := []struct {
		domain   string
		expected bool
	}{
		{"valid-domain", true},
		{"invalid@domain", false},
		{"Valid123", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.domain, func(t *testing.T) {
			isValid := IsValidDomain(testCase.domain)
			if isValid != testCase.expected {
				t.Errorf("Expected %s to be valid: %v, but got valid: %v", testCase.domain, testCase.expected, isValid)
			}
		})
	}
}
