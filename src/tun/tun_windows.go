//go:build windows
// +build windows

package tun

import (
	"errors"
	"log"
	"net/netip"
	"time"

	"github.com/RiV-chain/RiV-mesh/src/defaults"
	"golang.org/x/sys/windows"

	wgtun "golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/elevate"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// This is to catch Windows platforms

// Configures the TUN adapter with the correct IPv6 address and MTU.
func (tun *TunAdapter) setup(ifname string, addr string, mtu uint64) error {
	if ifname == "auto" {
		ifname = defaults.GetDefaults().DefaultIfName
	}
	return elevate.DoAsSystem(func() error {
		var err error
		var iface wgtun.Device
		var guid windows.GUID
		if guid, err = windows.GUIDFromString("{f1369c05-0344-40ed-a772-bfb4770abdd0}"); err != nil {
			return err
		}
		for i := 0; i < 15; i++ {
			if i > 0 {
				time.Sleep(time.Second)
				log.Printf("Retrying adapter creation after failure because system just booted (T+%v): %v", windows.DurationSinceBoot(), err)
			}
			iface, err = wgtun.CreateTUNWithRequestedGUID(ifname, &guid, int(mtu))
			if err == nil || windows.DurationSinceBoot() > time.Minute*10 {
				break
			}
		}
		if err != nil {
			return err
		}
		tun.iface = iface
		for i := 1; i < 10; i++ {
			if err = tun.setupAddress(addr); err != nil {
				tun.log.Errorln("Failed to set up TUN address:", err)
				log.Printf("waiting...")
				if i > 8 {
					return err
				} else {
					time.Sleep(time.Duration(2*i) * time.Second)
				}
			} else {
				break
			}
		}
		if err = tun.setupMTU(getSupportedMTU(mtu)); err != nil {
			tun.log.Errorln("Failed to set up TUN MTU:", err)
			return err
		}
		if mtu, err := iface.MTU(); err == nil {
			tun.mtu = uint64(mtu)
		}
		return nil
	})
}

// Sets the MTU of the TUN adapter.
func (tun *TunAdapter) setupMTU(mtu uint64) error {
	if tun.iface == nil || tun.Name() == "" {
		return errors.New("Can't configure MTU as TUN adapter is not present")
	}
	if intf, ok := tun.iface.(*wgtun.NativeTun); ok {
		luid := winipcfg.LUID(intf.LUID())
		ipfamily, err := luid.IPInterface(windows.AF_INET6)
		if err != nil {
			return err
		}

		ipfamily.NLMTU = uint32(mtu)
		intf.ForceMTU(int(ipfamily.NLMTU))
		ipfamily.UseAutomaticMetric = false
		ipfamily.Metric = 0
		ipfamily.DadTransmits = 0
		ipfamily.RouterDiscoveryBehavior = winipcfg.RouterDiscoveryDisabled

		if err := ipfamily.Set(); err != nil {
			return err
		}
	}

	return nil
}

// Sets the IPv6 address of the TUN adapter.
func (tun *TunAdapter) setupAddress(addr string) error {
	if tun.iface == nil || tun.Name() == "" {
		return errors.New("Can't configure IPv6 address as TUN adapter is not present")
	}
	if intf, ok := tun.iface.(*wgtun.NativeTun); ok {
		if address, err := netip.ParsePrefix(addr); err == nil {
			luid := winipcfg.LUID(intf.LUID())
			addresses := []netip.Prefix{address}

			err := luid.SetIPAddressesForFamily(windows.AF_INET6, addresses)
			if err == windows.ERROR_OBJECT_ALREADY_EXISTS {
				cleanupAddressesOnDisconnectedInterfaces(windows.AF_INET6, addresses)
				err = luid.SetIPAddressesForFamily(windows.AF_INET6, addresses)
			}
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		return errors.New("unable to get native TUN")
	}
	return nil
}

/*
 * cleanupAddressesOnDisconnectedInterfaces
 * SPDX-License-Identifier: MIT
 * Copyright (C) 2019 WireGuard LLC. All Rights Reserved.
 */
func cleanupAddressesOnDisconnectedInterfaces(family winipcfg.AddressFamily, addresses []netip.Prefix) {
	if len(addresses) == 0 {
		return
	}
	addrHash := make(map[netip.Addr]bool, len(addresses))
	for i := range addresses {
		addrHash[addresses[i].Addr()] = true
	}
	interfaces, err := winipcfg.GetAdaptersAddresses(family, winipcfg.GAAFlagDefault)
	if err != nil {
		return
	}
	for _, iface := range interfaces {
		if iface.OperStatus == winipcfg.IfOperStatusUp {
			continue
		}
		for address := iface.FirstUnicastAddress; address != nil; address = address.Next {
			if ip, _ := netip.AddrFromSlice(address.Address.IP()); addrHash[ip] {
				prefix := netip.PrefixFrom(ip, int(address.OnLinkPrefixLength))
				log.Printf("Cleaning up stale address %s from interface ‘%s’", prefix.String(), iface.FriendlyName())
				iface.LUID.DeleteIPAddress(prefix)
			}
		}
	}
}
