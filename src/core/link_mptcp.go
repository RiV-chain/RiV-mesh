package core

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/Arceliar/phony"
)

type linkMPTCP struct {
	phony.Inbox
	*links
	tcp        *linkTCP
	listener   *net.ListenConfig
	_listeners map[*Listener]context.CancelFunc
}

func (l *links) newLinkMPTCP(tcp *linkTCP) *linkMPTCP {
	lt := &linkMPTCP{
		links: l,
		tcp:   tcp,
		listener: &net.ListenConfig{
			Control:   tcp.tcpContext,
			KeepAlive: -1,
		},
		_listeners: map[*Listener]context.CancelFunc{},
	}
	lt.listener.Control = lt.tcp.tcpContext
	lt.listener.SetMultipathTCP(true)
	return lt
}

func (l *linkMPTCP) dial(ctx context.Context, url *url.URL, info linkInfo, options linkOptions) (net.Conn, error) {
	dialers, err := l.tcp.dialersFor(url, info)
	if err != nil {
		return nil, err
	}
	if len(dialers) == 0 {
		return nil, nil
	}
	for _, d := range dialers {
		var conn net.Conn
		d.dialer.SetMultipathTCP(true)
		if d.dialer.MultipathTCP() {
			l.core.log.Infof("Enabled MPTCP")
		} else {
			l.core.log.Infof("Enabled TCP")
		}
		conn, err = d.dialer.DialContext(ctx, "tcp", d.addr.String())
		if err != nil {
			l.core.log.Warnf("Failed to connect to %s: %s", d.addr, err)
			continue
		}
		return conn, nil
	}
	return nil, err
}

func (l *linkMPTCP) listen(ctx context.Context, url *url.URL, sintf string) (net.Listener, error) {
	hostport := url.Host
	if sintf != "" {
		if host, port, err := net.SplitHostPort(hostport); err == nil {
			hostport = fmt.Sprintf("[%s%%%s]:%s", host, sintf, port)
		}
	}
	return l.listener.Listen(ctx, "tcp", hostport)
}
