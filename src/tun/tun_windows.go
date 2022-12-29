//go:build windows
// +build windows

package tun

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"log"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"

	"golang.zx2c4.com/wintun"
	wgtun "golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/elevate"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

	"github.com/RiV-chain/RiV-mesh/src/defaults"
)

const (
	rateMeasurementGranularity = uint64((time.Second / 2) / time.Nanosecond)
	spinloopRateThreshold      = 800000000 / 8                                   // 800mbps
	spinloopDuration           = uint64(time.Millisecond / 80 / time.Nanosecond) // ~1gbit/s
)

type rateJuggler struct {
	current       atomic.Uint64
	nextByteCount atomic.Uint64
	nextStartTime atomic.Int64
	changing      atomic.Bool
}

type NativeTun struct {
	wt        *wintun.Adapter
	name      string
	handle    windows.Handle
	rate      rateJuggler
	session   wintun.Session
	readWait  windows.Handle
	events    chan wgtun.Event
	running   sync.WaitGroup
	closeOnce sync.Once
	close     atomic.Bool
	forcedMTU int
}

var (
	WintunTunnelType          = "WireGuard"
	WintunStaticRequestedGUID *windows.GUID
)

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
		iface, err = OpenTUNWithName(ifname, int(mtu))
		if err != nil {
			iface, err = CreateTUNWithRequestedGUID(ifname, &guid, int(mtu))
			if err != nil {
				wintun.Uninstall()
				iface, err = CreateTUNWithRequestedGUID(ifname, &guid, int(mtu))
				if err != nil {
					return err
				}
			}
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
	if intf, ok := tun.iface.(*NativeTun); ok {
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
	if intf, ok := tun.iface.(*NativeTun); ok {
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

//go:linkname procyield runtime.procyield
func procyield(cycles uint32)

//go:linkname nanotime runtime.nanotime
func nanotime() int64

// CreateTUN creates a Wintun interface with the given name. Should a Wintun
// interface with the same name exist, it is reused.
func CreateTUN(ifname string, mtu int) (wgtun.Device, error) {
	return CreateTUNWithRequestedGUID(ifname, WintunStaticRequestedGUID, mtu)
}

// CreateTUNWithRequestedGUID creates a Wintun interface with the given name and
// a requested GUID. Should a Wintun interface with the same name exist, it is reused.
func CreateTUNWithRequestedGUID(ifname string, requestedGUID *windows.GUID, mtu int) (wgtun.Device, error) {
	wt, err := wintun.CreateAdapter(ifname, WintunTunnelType, requestedGUID)
	if err != nil {
		return nil, fmt.Errorf("Error creating interface: %w", err)
	}

	forcedMTU := 1420
	if mtu > 0 {
		forcedMTU = mtu
	}

	tun := &NativeTun{
		wt:        wt,
		name:      ifname,
		handle:    windows.InvalidHandle,
		events:    make(chan wgtun.Event, 10),
		forcedMTU: forcedMTU,
	}

	tun.session, err = wt.StartSession(0x800000) // Ring capacity, 8 MiB
	if err != nil {
		tun.wt.Close()
		close(tun.events)
		return nil, fmt.Errorf("Error starting session: %w", err)
	}
	tun.readWait = tun.session.ReadWaitEvent()
	return tun, nil
}

// OpenTUNWithName opens a Wintun interface with the given name
func OpenTUNWithName(ifname string, mtu int) (wgtun.Device, error) {
	wt, err := wintun.OpenAdapter(ifname)
	if err != nil {
		return nil, fmt.Errorf("Error opening interface: %w", err)
	}

	forcedMTU := 1420
	if mtu > 0 {
		forcedMTU = mtu
	}

	tun := &NativeTun{
		wt:        wt,
		name:      ifname,
		handle:    windows.InvalidHandle,
		events:    make(chan wgtun.Event, 10),
		forcedMTU: forcedMTU,
	}

	tun.session, err = wt.StartSession(0x800000) // Ring capacity, 8 MiB
	if err != nil {
		tun.wt.Close()
		close(tun.events)
		return nil, fmt.Errorf("Error starting session: %w", err)
	}
	tun.readWait = tun.session.ReadWaitEvent()
	return tun, nil
}

func (tun *NativeTun) Name() (string, error) {
	return tun.name, nil
}

func (tun *NativeTun) File() *os.File {
	return nil
}

func (tun *NativeTun) Events() chan wgtun.Event {
	return tun.events
}

func (tun *NativeTun) Close() error {
	var err error
	tun.closeOnce.Do(func() {
		tun.close.Store(true)
		windows.SetEvent(tun.readWait)
		tun.running.Wait()
		tun.session.End()
		if tun.wt != nil {
			tun.wt.Close()
		}
		close(tun.events)
	})
	return err
}

func (tun *NativeTun) MTU() (int, error) {
	return tun.forcedMTU, nil
}

// TODO: This is a temporary hack. We really need to be monitoring the interface in real time and adapting to MTU changes.
func (tun *NativeTun) ForceMTU(mtu int) {
	update := tun.forcedMTU != mtu
	tun.forcedMTU = mtu
	if update {
		tun.events <- wgtun.EventMTUUpdate
	}
}

// Note: Read() and Write() assume the caller comes only from a single thread; there's no locking.

func (tun *NativeTun) Read(buff []byte, offset int) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
retry:
	if tun.close.Load() {
		return 0, os.ErrClosed
	}
	start := nanotime()
	shouldSpin := tun.rate.current.Load() >= spinloopRateThreshold && uint64(start-tun.rate.nextStartTime.Load()) <= rateMeasurementGranularity*2
	for {
		if tun.close.Load() {
			return 0, os.ErrClosed
		}
		packet, err := tun.session.ReceivePacket()
		switch err {
		case nil:
			packetSize := len(packet)
			copy(buff[offset:], packet)
			tun.session.ReleaseReceivePacket(packet)
			tun.rate.update(uint64(packetSize))
			return packetSize, nil
		case windows.ERROR_NO_MORE_ITEMS:
			if !shouldSpin || uint64(nanotime()-start) >= spinloopDuration {
				windows.WaitForSingleObject(tun.readWait, windows.INFINITE)
				goto retry
			}
			procyield(1)
			continue
		case windows.ERROR_HANDLE_EOF:
			return 0, os.ErrClosed
		case windows.ERROR_INVALID_DATA:
			return 0, errors.New("Send ring corrupt")
		}
		return 0, fmt.Errorf("Read failed: %w", err)
	}
}

func (tun *NativeTun) Flush() error {
	return nil
}

func (tun *NativeTun) Write(buff []byte, offset int) (int, error) {
	tun.running.Add(1)
	defer tun.running.Done()
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

	packetSize := len(buff) - offset
	tun.rate.update(uint64(packetSize))

	packet, err := tun.session.AllocateSendPacket(packetSize)
	if err == nil {
		copy(packet, buff[offset:])
		tun.session.SendPacket(packet)
		return packetSize, nil
	}
	switch err {
	case windows.ERROR_HANDLE_EOF:
		return 0, os.ErrClosed
	case windows.ERROR_BUFFER_OVERFLOW:
		return 0, nil // Dropping when ring is full.
	}
	return 0, fmt.Errorf("Write failed: %w", err)
}

// LUID returns Windows interface instance ID.
func (tun *NativeTun) LUID() uint64 {
	tun.running.Add(1)
	defer tun.running.Done()
	if tun.close.Load() {
		return 0
	}
	return tun.wt.LUID()
}

// RunningVersion returns the running version of the Wintun driver.
func (tun *NativeTun) RunningVersion() (version uint32, err error) {
	return wintun.RunningVersion()
}

func (rate *rateJuggler) update(packetLen uint64) {
	now := nanotime()
	total := rate.nextByteCount.Add(packetLen)
	period := uint64(now - rate.nextStartTime.Load())
	if period >= rateMeasurementGranularity {
		if !rate.changing.CompareAndSwap(false, true) {
			return
		}
		rate.nextStartTime.Store(now)
		rate.current.Store(total * uint64(time.Second/time.Nanosecond) / period)
		rate.nextByteCount.Store(0)
		rate.changing.Store(false)
	}
}
