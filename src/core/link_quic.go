package core

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"

	"github.com/Arceliar/phony"
	quicconn "github.com/vikulin/quic-conn"
)

type linkQUIC struct {
	phony.Inbox
	*links
	listener   *net.ListenConfig
	_listeners map[*Listener]context.CancelFunc
}

func (l *links) newLinkQUIC() *linkQUIC {
	lt := &linkQUIC{
		links: l,
		listener: &net.ListenConfig{
			KeepAlive: -1,
		},
		_listeners: map[*Listener]context.CancelFunc{},
	}
	return lt
}

func (l *linkQUIC) dial(url *url.URL, options linkOptions, sintf string) error {
	info := linkInfoFor("quic", sintf, strings.SplitN(url.Host, "%", 2)[0])
	if l.links.isConnectedTo(info) {
		return nil
	}
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic"},
	}
	conn, err := quicconn.Dial(url.Host, tlsConf)
	if err != nil {
		return err
	}
	dial := &linkDial{
		url:   url,
		sintf: sintf,
	}
	return l.handler(dial, url.String(), info, conn, options, false, false)
}

func (l *linkQUIC) listen(url *url.URL, sintf string) (*Listener, error) {
	//_, cancel := context.WithCancel(l.core.ctx)
	/*
		hostport := url.Host
		if sintf != "" {
			if host, port, err := net.SplitHostPort(hostport); err == nil {
				hostport = fmt.Sprintf("[%s%%%s]:%s", host, sintf, port)
			}
		}
	*/
	tlsConf := generateTLSConfig()

	listener, err := quicconn.Listen("udp", url.Host, tlsConf)

	if err != nil {
		//cancel()
		return nil, err
	}

	entry := &Listener{
		Listener: listener,
		closed:   make(chan struct{}),
	}
	//phony.Block(l, func() {
	//	l._listeners[entry] = cancel
	//})
	l.core.log.Printf("QUIC listener started on %s", listener.Addr())
	go func() {
		defer phony.Block(l, func() {
			delete(l._listeners, entry)
		})
		for {
			conn, err := listener.Accept()
			if err != nil {
				//cancel()
				break
			}
			raddr := conn.RemoteAddr().(*net.UDPAddr)
			if err != nil {
				break
			}
			name := fmt.Sprintf("quic://%s", raddr)
			info := linkInfoFor("quic", sintf, raddr.String())
			if err = l.handler(nil, name, info, conn, linkOptionsForListener(url), true, raddr.IP.IsLinkLocalUnicast()); err != nil {
				l.core.log.Errorln("Failed to create inbound link:", err)
			}
		}
		_ = listener.Close()
		close(entry.closed)
		l.core.log.Printf("QUIC listener stopped on %s", listener.Addr())
	}()
	return entry, nil
}

func (l *linkQUIC) handler(dial *linkDial, name string, info linkInfo, conn net.Conn, options linkOptions, incoming, force bool) error {
	return l.links.create(
		conn,     // connection
		dial,     // connection URL
		name,     // connection name
		info,     // connection info
		incoming, // not incoming
		force,    // not forced
		options,  // connection options
	)
}

// QUIC infrastructure
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic"},
	}
}
