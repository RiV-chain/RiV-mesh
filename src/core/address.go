// Package address contains the types used by mesh to represent IPv6 addresses or prefixes, as well as functions for working with these types.
// Of particular importance are the functions used to derive addresses or subnets from a NodeID, or to get the NodeID and bitmask of the bits visible from an address, which is needed for DHT searches.
package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/eknkc/basex"

	iwt "github.com/Arceliar/ironwood/types"
)

const domainNameCharacters string = "0123456789abcdefghijklmnopqrstuvwxyz-"

// Address represents an IPv6 address in the mesh address range.
type Address [16]byte

// Subnet represents an IPv6 /64 subnet in the mesh subnet range.
type Subnet [8]byte

// GetPrefix returns the address prefix used by mesh.
// The current implementation requires this to be a multiple of 8 bits + 7 bits.
// The 8th bit of the last byte is used to signal nodes (0) or /64 prefixes (1).
// Nodes that configure this differently will be unable to communicate with each other using IP packets, though routing and the DHT machinery *should* still work.
func (c *Core) GetPrefix() [1]byte {
	p, err := hex.DecodeString(c.config.networkdomain.Prefix)
	if err != nil {
		panic(err)
	}
	var prefix [1]byte
	copy(prefix[:], p[:1])
	return prefix
}

// IsValid returns true if an address falls within the range used by nodes in the network.
func (c *Core) IsValidAddress(a Address) bool {
	prefix := c.GetPrefix()
	for idx := range prefix {
		if a[idx] != prefix[idx] {
			return false
		}
	}
	return true
}

// IsValid returns true if a prefix falls within the range usable by the network.
func (c *Core) IsValidSubnet(s Subnet) bool {
	prefix := c.GetPrefix()
	l := len(prefix)
	for idx := range prefix[:l-1] {
		if s[idx] != prefix[idx] {
			return false
		}
	}
	return s[l-1] == prefix[l-1]|0x01
}

// AddrForDomain takes a Domain as an argument and returns an *Address.
// This function returns nil if the key length is not ed25519.PublicKeySize.
// This address begins with the contents of GetPrefix(), with the last bit set to 0 to indicate an address.
// The following 8 bits are set to the number of leading 1 bits in the bitwise inverse of the public key.
// The bitwise inverse of the Domain name, excluding the leading 1 bits and the first leading 0 bit, is truncated to the appropriate length and makes up the remainder of the address.
func (c *Core) AddrForDomain(domain iwt.Domain) *Address {
	// 128 bit address
	// Begins with prefix
	// Next bit is a 0
	// Next 7 bits, interpreted as a uint, are # of leading 1s in the NodeID
	// Leading 1s and first leading 0 of the NodeID are truncated off
	// The rest is appended to the IPv6 address (truncated to 128 bits total)
	if len(domain.Key) != ed25519.PublicKeySize {
		return nil
	}
	addr, err := encodeToIPv6(c.GetPrefix(), domain.Name)
	if err != nil {
		c.log.Errorln(err)
		return nil
	}
	return &addr
}

// SubnetForDomain takes a Domain as an argument and returns a *Subnet.
// This function returns nil if the key length is not ed25519.PublicKeySize.
// The subnet begins with the address prefix, with the last bit set to 1 to indicate a prefix.
// The following 8 bits are set to the number of leading 1 bits in the bitwise inverse of the key.
// The bitwise inverse of the Domain name bytes, excluding the leading 1 bits and the first leading 0 bit, is truncated to the appropriate length and makes up the remainder of the subnet.
func (c *Core) SubnetForDomain(domain iwt.Domain) *Subnet {
	// Exactly as the address version, with two exceptions:
	//  1) The first bit after the fixed prefix is a 1 instead of a 0
	//  2) It's truncated to a subnet prefix length instead of 128 bits
	addr := c.AddrForDomain(domain)
	if addr == nil {
		return nil
	}
	var snet Subnet
	copy(snet[:], addr[:])
	snet[len(c.GetPrefix())-1] |= 0x01
	return &snet
}

// Returns the partial Domain for the Address.
// This is used for domain lookup.
func (c *Core) GetAddressDomain(a Address) iwt.Domain {
	name, err := decodeIPv6(a)
	if err != nil {
		return iwt.Domain{}
	}
	//zero filled byte array here
	var bytes [ed25519.PublicKeySize]byte

	return iwt.NewDomain(string(name), bytes[:])
}

// Returns the partial Domain for the Subnet.
// This is used for domain lookup.
func (c *Core) GetSubnetDomain(s Subnet) iwt.Domain {
	var addr Address
	copy(addr[:], s[:])
	return c.GetAddressDomain(addr)
}

func encodeToIPv6(prefix [1]byte, name []byte) (Address, error) {
	str := string(truncateTrailingZeros(name))
	if len(str) > 23 {
		return Address{}, fmt.Errorf("input data is too long for an IPv6 address")
	}
	encoder, err := basex.NewEncoding(domainNameCharacters)
	if err != nil {
		return Address{}, err
	}
	var ipv6Bytes Address
	copy(ipv6Bytes[:], prefix[:])
	decoded, err := encoder.Decode(str)
	if err != nil {
		return Address{}, errors.New("Base37 decode error in string:" + str)
	}
	copy(ipv6Bytes[1:], decoded)
	return ipv6Bytes, nil
}

func decodeIPv6(ipv6 Address) ([]byte, error) {
	encoder, err := basex.NewEncoding(domainNameCharacters)
	if err != nil {
		return nil, err
	}
	encodedData := truncateTrailingZeros(ipv6[1:])
	dst := []byte(encoder.Encode(encodedData))
	return dst, nil
}

func truncateTrailingZeros(data []byte) []byte {
	length := len(data)
	for length > 0 && data[length-1] == 0 {
		length--
	}
	return data[:length]
}
