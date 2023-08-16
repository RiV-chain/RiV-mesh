package dnsapi

import (
	"context"
	"net"
	"os"
	"strings"

	"github.com/Arceliar/ironwood/types"
	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/gologme/log"
	"github.com/miekg/dns"
	"github.com/mikispag/dns-over-tls-forwarder/proxy"
)

type DnsServerCfg struct {
	Core            *core.Core
	Log             core.Logger
	ListenAddress   string
	upstreamServers []string
	Tld             string
}

type DnsServer struct {
	server *proxy.Server
	DnsServerCfg
}

func NewDnsServer(cfg DnsServerCfg) (*DnsServer, error) {
	s := &DnsServer{
		server:       proxy.NewServer(0, false, cfg.upstreamServers...),
		DnsServerCfg: cfg,
	}

	return s, nil
}

func (s *DnsServer) Run() {
	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-sigs
		cancel()
	}()
	go func() {
		s.Log.Errorln(s.server.RunWithHandle(ctx, s.ListenAddress, s.ServeDNS))
	}()
}

func (s *DnsServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	msg.Compress = false
	for _, q := range r.Question {
		if q.Qtype == dns.TypeAAAA && strings.HasSuffix(q.Name, s.Tld) {
			name := strings.TrimSuffix(q.Name, s.Tld)
			aaaaRecord := &dns.AAAA{
				Hdr:  dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60},
				AAAA: net.ParseIP(net.IP(s.Core.AddrForDomain(types.NewDomain(name, s.Core.PublicKey()))[:]).String()),
			}
			msg.Answer = append(msg.Answer, aaaaRecord)
		}
	}
	if len(msg.Answer) > 0 {
		err := w.WriteMsg(&msg)
		s.Log.Errorln(err)
	} else {
		inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
		s.Log.Debugf("Question from %s: %q", inboundIP, r.Question[0])
		m := s.server.GetAnswer(r)
		if m == nil {
			dns.HandleFailed(w, r)
			return
		}
		if err := w.WriteMsg(m); err != nil {
			log.Warnf("Write message failed, message: %v, error: %v", m, err)
		}
	}
}
