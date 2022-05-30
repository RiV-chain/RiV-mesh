package core

// This sends packets to peers using TCP as a transport
// It's generally better tested than the UDP implementation
// Using it regularly is insane, but I find TCP easier to test/debug with it
// Updating and optimizing the UDP version is a higher priority

// TODO:
//  Something needs to make sure we're getting *valid* packets
//  Could be used to DoS (connect, give someone else's keys, spew garbage)
//  I guess the "peer" part should watch for link packets, disconnect?

// TCP connections start with a metadata exchange.
//  It involves exchanging version numbers and crypto keys
//  See version.go for version metadata format

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"github.com/RiV-chain/RiV-mesh/src/address"
	//"github.com/RiV-chain/RiV-mesh/src/util"
	kcpconn "github.com/xtaci/kcp-go/v5"
)

const default_timeout = 6 * time.Second

// The TCP listener and information about active TCP connections, to avoid duplication.
type tcp struct {
	links     *links
	waitgroup sync.WaitGroup
	mutex     sync.Mutex // Protecting the below
	listeners map[string]*TcpListener
	calls     map[string]struct{}
	conns     map[linkInfo](chan struct{})
	tls       tcptls
}

// TcpListener is a stoppable TCP listener interface. These are typically
// returned from calls to the ListenTCP() function and are also used internally
// to represent listeners created by the "Listen" configuration option and for
// multicast interfaces.
type TcpListener struct {
	Listener net.Listener
	opts     tcpOptions
	stop     chan struct{}
}

type TcpUpgrade struct {
	upgrade func(c net.Conn, o *tcpOptions) (net.Conn, error)
	name    string
}

type tcpOptions struct {
	linkOptions
	upgrade        *TcpUpgrade
	socksProxyAddr string
	socksProxyAuth *proxy.Auth
	socksPeerAddr  string
	tlsSNI         string
}

func (l *TcpListener) Stop() {
	defer func() { _ = recover() }()
	close(l.stop)
}

// Wrapper function to set additional options for specific connection types.
func (t *tcp) setExtraOptions(c net.Conn) {
	switch sock := c.(type) {
	case *net.TCPConn:
		_ = sock.SetNoDelay(true)
	// TODO something for socks5
	default:
	}
}

// Returns the address of the listener.
func (t *tcp) getAddr() *net.TCPAddr {
	// TODO: Fix this, because this will currently only give a single address
	// to multicast.go, which obviously is not great, but right now multicast.go
	// doesn't have the ability to send more than one address in a packet either
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, l := range t.listeners {
		return l.Listener.Addr().(*net.TCPAddr)
	}
	return nil
}

// Initializes the struct.
func (t *tcp) init(l *links) error {
	t.links = l
	t.tls.init(t)
	t.mutex.Lock()
	t.calls = make(map[string]struct{})
	t.conns = make(map[linkInfo](chan struct{}))
	t.listeners = make(map[string]*TcpListener)
	t.mutex.Unlock()

	t.links.core.config.RLock()
	defer t.links.core.config.RUnlock()
	for _, listenaddr := range t.links.core.config.Listen {
		u, err := url.Parse(listenaddr)
		if err != nil {
			t.links.core.log.Errorln("Failed to parse listener: listener", listenaddr, "is not correctly formatted, ignoring")
		}
		if _, err := t.listenURL(u, ""); err != nil {
			return err
		}
	}

	return nil
}

func (t *tcp) stop() error {
	t.mutex.Lock()
	for _, listener := range t.listeners {
		listener.Stop()
	}
	t.mutex.Unlock()
	t.waitgroup.Wait()
	return nil
}

func (t *tcp) listenURL(u *url.URL, sintf string) (*TcpListener, error) {
	var listener *TcpListener
	var err error
	var urlstring string
	//reconstruct URL here
	
	if len(sintf) != 0 {
		host, port, err := net.SplitHostPort(u.Host)
		if err == nil {
			urlstring = fmt.Sprintf("%s://[%s%%25%s]:%s", u.Scheme, host, url.QueryEscape(sintf), port)
		}
		u, err = url.Parse(urlstring)
		if err != nil {
			t.links.core.log.Errorln("Failed to parse listener: url", urlstring, "is not correctly formatted, ignoring")
			return nil, err
		}
	}
	switch u.Scheme {
	case "tcp":
		listener, err = t.listen(u, nil)
	case "tls":
		listener, err = t.listen(u, t.tls.forListener)
	case "kcp":
		listener, err = t.listenKcp(u, t.tls)
	default:
		t.links.core.log.Errorln("Failed to add listener: listener", u.String(), "is not correctly formatted, ignoring")
	}
	return listener, err
}

func (t *tcp) listen(u *url.URL, upgrade *TcpUpgrade) (*TcpListener, error) {
	var err error

	ctx := t.links.core.ctx
	lc := net.ListenConfig{
		Control: t.tcpContext,
	}
	listener, err := lc.Listen(ctx, "tcp", u.Host)
	if err == nil {
		l := TcpListener{
			Listener: listener,
			opts:     tcpOptions{
				upgrade: upgrade,
			},
			stop:     make(chan struct{}),
		}
		t.waitgroup.Add(1)
		go t.listener(&l, u)
		return &l, nil
	} else {
		t.links.core.log.Errorln("Failed start listener: ", u.Host)
	}

	return nil, err
}

func (t *tcp) listenKcp(u *url.URL, _ tcptls) (*TcpListener, error) {
	//keep tls for future encryption
	var err error
	listener, err := kcpconn.Listen(u.Host)
	if err == nil {
		//update proto here?
		//tls.forListener.name = "quic"
		l := TcpListener{
			Listener: listener,
			opts:     tcpOptions{
				upgrade: nil,
			},
			stop:     make(chan struct{}),
		}
		t.waitgroup.Add(1)
		go t.listener(&l, u)
		return &l, nil
	} else {
		t.links.core.log.Errorln("Failed start listener: ", u.Host)
	}

	return nil, err
}

// Runs the listener, which spawns off goroutines for incoming connections.
func (t *tcp) listener(l *TcpListener, u *url.URL) {
	defer t.waitgroup.Done()
	if l == nil {
		return
	}
	// Track the listener so that we can find it again in future
	t.mutex.Lock()
	if _, isIn := t.listeners[u.Host]; isIn {
		t.mutex.Unlock()
		l.Listener.Close()
		return
	}
	t.listeners[u.Host] = l
	t.mutex.Unlock()
	// And here we go!
	defer func() {
		t.links.core.log.Infoln("Stopping", u.String(), "listener on:", l.Listener.Addr().String())
		l.Listener.Close()
		t.mutex.Lock()
		delete(t.listeners, u.Host)
		t.mutex.Unlock()
	}()
	t.links.core.log.Infoln("Listening for", u.String(), "on:", l.Listener.Addr().String())
	go func() {
		<-l.stop
		l.Listener.Close()
	}()
	defer l.Stop()
	for {
		t.links.core.log.Infoln("Accepting listener for", u.String())
		sock, err := l.Listener.Accept()
		t.links.core.log.Infoln("Accepted listener for", sock)
		if err != nil {
			t.links.core.log.Errorln("Failed to accept connection:", err)
			select {
			case <-l.stop:
				return
			default:
			}
			time.Sleep(time.Second) // So we don't busy loop
			continue
		}
		t.waitgroup.Add(1)
		options := l.opts
		go t.handler(sock, true, options)
	}
}

// Checks if we already are calling this address
func (t *tcp) startCalling(saddr string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	_, isIn := t.calls[saddr]
	t.calls[saddr] = struct{}{}
	return !isIn
}

// Checks if a connection already exists.
// If not, it adds it to the list of active outgoing calls (to block future attempts) and dials the address.
// If the dial is successful, it launches the handler.
// When finished, it removes the outgoing call, so reconnection attempts can be made later.
// This all happens in a separate goroutine that it spawns.
func (t *tcp) call(u *url.URL, options tcpOptions, sintf string) {
	go func() {
		t.links.core.log.Debugf("Started dial call for %s ", u)
		callname := u.Host
		callproto := strings.ToUpper(u.Scheme)
		if options.upgrade != nil {
			callproto = strings.ToUpper(options.upgrade.name)
		}
		if sintf != "" {
			callname = fmt.Sprintf("%s/%s/%s", callproto, u.Host, sintf)
		}
		if !t.startCalling(callname) {
			return
		}
		defer func() {
			// Block new calls for a little while, to mitigate livelock scenarios
			rand.Seed(time.Now().UnixNano())
			delay := default_timeout + time.Duration(rand.Intn(10000))*time.Millisecond
			time.Sleep(delay)
			t.mutex.Lock()
			delete(t.calls, callname)
			t.mutex.Unlock()
		}()
		var conn net.Conn
		var err error
		if options.socksProxyAddr != "" {
			if sintf != "" {
				return
			}
			dialerdst, er := net.ResolveTCPAddr("tcp", options.socksProxyAddr)
			if er != nil {
				return
			}
			var dialer proxy.Dialer
			dialer, err = proxy.SOCKS5("tcp", dialerdst.String(), options.socksProxyAuth, proxy.Direct)
			if err != nil {
				return
			}
			ctx, done := context.WithTimeout(t.links.core.ctx, default_timeout)
			pathtokens := strings.Split(strings.Trim(u.Path, "/"), "/")
			conn, err = dialer.(proxy.ContextDialer).DialContext(ctx, "tcp", pathtokens[0])
			done()
			if err != nil {
				return
			}
			t.waitgroup.Add(1)
			options.socksPeerAddr = u.Host
			if ch := t.handler(conn, false, options); ch != nil {
				<-ch
			}
		} else {
			t.links.core.log.Debugf("Resolving IP %s ", u.Host)
			host, port, _ := net.SplitHostPort(u.Host)
			if err != nil {
				t.links.core.log.Errorln("URL host:port parsing failed:", err.Error())
				return
			}
			dst, err := net.ResolveIPAddr("ip", host)
			if err != nil {
				t.links.core.log.Errorln("Resolving failed:", err.Error())
				return
			}
			if dst.IP.IsLinkLocalUnicast() {
				dst.Zone = sintf
				if dst.Zone == "" {
					return
				}
			}
			dialer := net.Dialer{
				Control: t.tcpContext,
			}
			t.links.core.log.Debugf("Dial created")
			if sintf != "" {
				dialer.Control = t.getControl(sintf)
				ief, err := net.InterfaceByName(sintf)
				if err != nil {
					return
				}
				if ief.Flags&net.FlagUp == 0 {
					return
				}
				addrs, err := ief.Addrs()
				if err == nil {
					for addrindex, addr := range addrs {
						src, _, err := net.ParseCIDR(addr.String())
						if err != nil {
							continue
						}
						if src.Equal(dst.IP) {
							continue
						}
						if !src.IsGlobalUnicast() && !src.IsLinkLocalUnicast() {
							continue
						}
						bothglobal := src.IsGlobalUnicast() == dst.IP.IsGlobalUnicast()
						bothlinklocal := src.IsLinkLocalUnicast() == dst.IP.IsLinkLocalUnicast()
						if !bothglobal && !bothlinklocal {
							continue
						}
						if (src.To4() != nil) != (dst.IP.To4() != nil) {
							continue
						}
						if bothglobal || bothlinklocal || addrindex == len(addrs)-1 {
							dialer.LocalAddr = &net.TCPAddr{
								IP:   src,
								Port: 0,
								Zone: sintf,
							}
							break
						}
					}
					if dialer.LocalAddr == nil {
						return
					}
				}
			}
			ctx, done := context.WithTimeout(t.links.core.ctx, default_timeout)
			t.links.core.log.Debugf("Starting dial contect for %s", dst.String()+":"+port)
			switch u.Scheme {
			case "tcp":
				conn, err = dialer.DialContext(ctx, "tcp", dst.String()+":"+port)
			case "tls":
				conn, err = dialer.DialContext(ctx, "tcp", dst.String()+":"+port)
			case "kcp":
				conn, err = kcpconn.Dial(dst.String()+":"+port)
			default:
				t.links.core.log.Errorln("Unknown schema:", u.String(), " is not correctly formatted, ignoring")
				return
			}
			
			done()
			if err != nil {
				t.links.core.log.Debugf("Failed to dial %s: %s", callproto, err)
				return
			}
			t.waitgroup.Add(1)
			if ch := t.handler(conn, false, options); ch != nil {
				<-ch
			}
		}
	}()
}

func (t *tcp) handler(sock net.Conn, incoming bool, options tcpOptions) chan struct{} {
	t.links.core.log.Debugf("Started dial handler for %s ", sock)
	defer t.waitgroup.Done() // Happens after sock.close
	defer sock.Close()
	t.setExtraOptions(sock)
	var upgraded bool
	if options.upgrade != nil {
		var err error
		if sock, err = options.upgrade.upgrade(sock, &options); err != nil {
			t.links.core.log.Errorln("TCP handler upgrade failed:", err)
			return nil
		}
		upgraded = true
	}
	var name, proto, local, remote string
	if options.socksProxyAddr != "" {
		name = "socks://" + sock.RemoteAddr().String() + "/" + options.socksPeerAddr
		proto = "socks"
		local, _, _ = net.SplitHostPort(sock.LocalAddr().String())
		remote, _, _ = net.SplitHostPort(options.socksPeerAddr)
	} else {
		if upgraded {
			proto = options.upgrade.name
			name = proto + "://" + sock.RemoteAddr().String()
		} else {
			proto = "tcp"
			name = proto + "://" + sock.RemoteAddr().String()
		}
		local, _, _ = net.SplitHostPort(sock.LocalAddr().String())
		remote, _, _ = net.SplitHostPort(sock.RemoteAddr().String())
	}
	localIP := net.ParseIP(local)
	if localIP = localIP.To16(); localIP != nil {
		var laddr address.Address
		var lsubnet address.Subnet
		copy(laddr[:], localIP)
		copy(lsubnet[:], localIP)
		if laddr.IsValid() || lsubnet.IsValid() {
			// The local address is with the network address/prefix range
			// This would route ygg over ygg, which we don't want
			// FIXME ideally this check should happen outside of the core library
			//  Maybe dial/listen at the application level
			//  Then pass a net.Conn to the core library (after these kinds of checks are done)
			t.links.core.log.Debugln("Dropping ygg-tunneled connection", local, remote)
			return nil
		}
	}
	force := net.ParseIP(strings.Split(remote, "%")[0]).IsLinkLocalUnicast()
	link, err := t.links.create(sock, name, proto, local, remote, incoming, force, options.linkOptions)
	if err != nil {
		t.links.core.log.Println(err)
		panic(err)
	}
	t.links.core.log.Debugln("DEBUG: starting handler for", name)
	ch, err := link.handler()
	t.links.core.log.Debugln("DEBUG: stopped handler for", name, err)
	return ch
}
