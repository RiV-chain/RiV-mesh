//go:build android
// +build android

package core

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"github.com/getlantern/multipath"

	"github.com/Arceliar/phony"
)

func (l *linkMPATH) connFor(url *url.URL, sinterfaces string) (net.Conn, error) {
	//Peer url has following format: mpath://host-1:port-1/host-2:port-2.../host-n:port-n
	hosts := strings.Split(url.String(), "/")[2:]
	remoteTargets := make([]net.Addr, 0)
	for _, host := range hosts {
		dst, err := net.ResolveTCPAddr("tcp", host)
		if err != nil {
			l.core.log.Errorln("could not resolve host ", dst.String())
			continue
		}
		if dst.IP.IsLinkLocalUnicast() {
			dst.Zone = sinterfaces
			if dst.Zone == "" {
				l.core.log.Errorln("link-local address requires a zone in ", dst.String())
				continue
			}
		}
		remoteTargets = append(remoteTargets, dst)
	}

	if len(remoteTargets) == 0 {
		return nil, fmt.Errorf("no valid target hosts given")
	}

	dialers := make([]multipath.Dialer, 0)
	trackers := make([]multipath.StatsTracker, 0)
	if sinterfaces != "" {
		sintfarray := strings.Split(sinterfaces, ",")
		for _, dst := range remoteTargets {
			for _, sintf := range sintfarray { 
				src, err := net.ParseIP(sintf)
				if err != nil {
					l.core.log.Errorln("interface %s address incorrect: %w", sintf, err)
					continue
				}
				dstIp := dst.(*net.TCPAddr).IP
				for addrindex, addr := range addrs {
					if !src.IsGlobalUnicast() && !src.IsLinkLocalUnicast() {
						continue
					}
					bothglobal := src.IsGlobalUnicast() == dstIp.IsGlobalUnicast()
					bothlinklocal := src.IsLinkLocalUnicast() == dstIp.IsLinkLocalUnicast()
					if !bothglobal && !bothlinklocal {
						continue
					}
					if (src.To4() != nil) != (dstIp.To4() != nil) {
						continue
					}
					if bothglobal || bothlinklocal || addrindex == len(addrs)-1 {
						td := newOutboundDialer(src, dst)
						dialers = append(dialers, td)
						trackers = append(trackers, multipath.NullTracker{})
						l.core.log.Printf("added outbound dialer for %s->%s", src.String(), dst.String())
						break
					}
				}
			}
		}
	} else {
		star := net.ParseIP("0.0.0.0")
		for _, dst := range remoteTargets {
			td := newOutboundDialer(star, dst)
			dialers = append(dialers, td)
			trackers = append(trackers, multipath.NullTracker{})
			l.core.log.Printf("added outbound dialer for %s", dst.String())
		}
	}
	if len(dialers) == 0 {
		return nil, fmt.Errorf("no suitable source address found on interface %q", sinterfaces)
	}
	dialer := multipath.NewDialer("mpath", dialers)
	//conn, err := dialer.DialContext(l.core.ctx, "tcp", remoteTargets[0].String())
	conn, err := dialer.DialContext(l.core.ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
