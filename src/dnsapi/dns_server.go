package dnsapi

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
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
			var bytes [ed25519.PublicKeySize]byte
			addr := s.Core.AddrForDomain(types.NewDomain(name, bytes[:]))
			aaaaRecord := &dns.AAAA{
				Hdr:  dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60},
				AAAA: net.ParseIP(net.IP(addr[:]).String()),
			}
			msg.Answer = append(msg.Answer, aaaaRecord)
		} else if q.Qtype == dns.TypePTR && q.Qclass == dns.ClassINET && strings.HasSuffix(q.Name, ".c.f.ip6.arpa.") {
			ptr, _ := dns.ReverseAddr(s.Core.Address().String())
			if q.Name == ptr {
				msg.Answer = append(msg.Answer, createPTRRecord(q.Name, s.Domain+s.DnsServerCfg.Tld))
			} else {
				//send PTR request to another server here
				resolver := &net.Resolver{}
				dnsServer, err := ptrToIPv6(q.Name)
				if err != nil {
					msg.SetRcode(r, dns.RcodeFormatError)
				} else {
					resp, err := lookupDNSRecord(resolver, dnsServer, q)
					if err != nil {
						msg.SetRcode(r, dns.RcodeFormatError)
					} else {
						msg.Answer = append(msg.Answer, resp.Answer...)
					}
				}
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

func ptrToIPv6(arpa string) (string, error) {
	mainPtr := arpa[:len(arpa)-9]
	pieces := strings.Split(mainPtr, ".")
	reversePieces := make([]string, len(pieces))
	for i := len(pieces) - 1; i >= 0; i-- {
		reversePieces[len(pieces)-1-i] = pieces[i]
	}
	hexString := strings.Join(reversePieces, "")
	ipBytes, err := hex.DecodeString(hexString)
	if err != nil {
		return "", err
	}
	ipv6Addr := net.IP(ipBytes).String()
	return ipv6Addr, nil
}

func lookupDNSRecord(resolver *net.Resolver, dnsServer string, q dns.Question) (r *dns.Msg, err error) {
	// Create a DNS client with custom DNS server
	client := &dns.Client{}

	msg := new(dns.Msg)
	msg.SetQuestion(q.Name, dns.TypePTR)

	// Send the DNS query
	ipv6Addr := net.ParseIP(dnsServer)
	serverAddr := &net.UDPAddr{IP: ipv6Addr, Port: 53}
	resp, _, err := client.ExchangeContext(context.Background(), msg, serverAddr.String())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

/*
func createPTRRecordFromResponse(msg dns.Msg) dns.RR {
	rr := new(dns.PTR)
	rr.Hdr = dns.RR_Header{Name: msg., Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 3600}
	rr.Ptr = target
	return rr
}
*/

func createPTRRecord(ptrName, target string) dns.RR {
	rr := new(dns.PTR)
	rr.Hdr = dns.RR_Header{Name: ptrName, Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 3600}
	rr.Ptr = target
	return rr
}
