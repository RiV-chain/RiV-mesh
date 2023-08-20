package dnsapi

import (
	"context"
	"net"
	"strings"

	"github.com/gologme/log"

	"github.com/Arceliar/ironwood/types"
	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/miekg/dns"
	"github.com/mikispag/dns-over-tls-forwarder/proxy"
)

type DnsServerCfg struct {
	Core            *core.Core
	Log             core.Logger
	ListenAddress   string
	UpstreamServers []string
	Tld             string
}

type DnsServer struct {
	server *proxy.Server
	DnsServerCfg
	Domain string
}

func NewDnsServer(domain string, cfg DnsServerCfg) (*DnsServer, error) {
	mux := dns.NewServeMux()
	s := &DnsServer{
		server:       proxy.NewServer(mux, cfg.Log.(*log.Logger), 0, false, cfg.ListenAddress, cfg.UpstreamServers...),
		DnsServerCfg: cfg,
		Domain:       domain,
	}
	mux.HandleFunc(".", s.ServeDNS)
	return s, nil
}

func (s *DnsServer) Run() {
	go func() {
		s.Log.Errorln(s.server.Run(context.Background()))
	}()
}

// Shutdown http server
func (s *DnsServer) Shutdown() error {
	err := s.server.Shutdown(context.Background())
	s.Log.Infof("Stop DNS server")
	return err
}

func (s *DnsServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := &dns.Msg{}
	msg.SetReply(r)
	msg.Compress = false
	for _, q := range r.Question {
		if (q.Qtype == dns.TypeA || q.Qtype == dns.TypeAAAA) && strings.HasSuffix(q.Name, s.Tld) {
			name := strings.TrimSuffix(q.Name, s.Tld)
			aaaaRecord := &dns.AAAA{
				Hdr:  dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60},
				AAAA: net.ParseIP(net.IP(s.Core.AddrForDomain(types.NewDomain(name, nil))[:]).String()),
			}
			msg.Answer = append(msg.Answer, aaaaRecord)
		} else if q.Qtype == dns.TypePTR && q.Qclass == dns.ClassINET && strings.HasPrefix(q.Name, "ip6.arpa.") {
			ip := extractIPv6FromPTR(q.Name)
			localIp := net.ParseIP(s.Core.Address().String())
			if ip.Equal(localIp) {
				msg.Answer = append(msg.Answer, createPTRRecord(q.Name, s.Domain+s.DnsServerCfg.Tld))
			}
		}
	}
	if len(msg.Answer) > 0 {
		if err := w.WriteMsg(msg); err != nil {
			s.Log.Warnf("Write message failed, message: %v, error: %v", msg, err)
		}
	} else {
		inboundIP, _, _ := net.SplitHostPort(w.RemoteAddr().String())
		s.Log.Debugf("Question from %s: %q", inboundIP, r.Question[0])
		m := s.server.GetAnswer(r)
		if m == nil {
			dns.HandleFailed(w, r)
			return
		}
		if err := w.WriteMsg(m); err != nil {
			s.Log.Warnf("Write message failed, message: %v, error: %v", m, err)
		}
	}
}

func extractIPv6FromPTR(ptrName string) net.IP {
	parts := strings.Split(ptrName, ".")
	if len(parts) == 6 {
		ipv6Str := parts[0] + parts[1] + parts[2] + parts[3]
		ipv6 := net.ParseIP(ipv6Str)
		return ipv6
	}
	return nil
}

func createPTRRecord(ptrName, target string) dns.RR {
	rr := new(dns.PTR)
	rr.Hdr = dns.RR_Header{Name: ptrName, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 3600}
	rr.Ptr = target
	return rr
}
