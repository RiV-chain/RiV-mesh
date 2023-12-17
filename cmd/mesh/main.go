package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Arceliar/ironwood/types"
	"github.com/gologme/log"
	gsyslog "github.com/hashicorp/go-syslog"
	"github.com/hjson/hjson-go"
	"github.com/kardianos/minwinsvc"

	"github.com/RiV-chain/RiV-mesh/src/config"
	"github.com/RiV-chain/RiV-mesh/src/dnsapi"

	"github.com/RiV-chain/RiV-mesh/src/core"
	"github.com/RiV-chain/RiV-mesh/src/multicast"
	"github.com/RiV-chain/RiV-mesh/src/restapi"
	"github.com/RiV-chain/RiV-mesh/src/tun"
	"github.com/RiV-chain/RiV-mesh/src/version"
)

type node struct {
	core        *core.Core
	tun         *tun.TunAdapter
	multicast   *multicast.Multicast
	rest_server *restapi.RestServer
	dns_server  *dnsapi.DnsServer
}

func setLogLevel(loglevel string, logger *log.Logger) {
	levels := [...]string{"error", "warn", "info", "debug", "trace"}
	loglevel = strings.ToLower(loglevel)

	contains := func() bool {
		for _, l := range levels {
			if l == loglevel {
				return true
			}
		}
		return false
	}

	if !contains() { // set default log level
		logger.Infoln("Loglevel parse failed. Set default level(info)")
		loglevel = "info"
	}

	for _, l := range levels {
		logger.EnableLevel(l)
		if l == loglevel {
			break
		}
	}
}

// The main function is responsible for configuring and starting RiV-mesh
func run(sigCh chan os.Signal) {
	genconf := flag.Bool("genconf", false, "print a new config to stdout")
	useconf := flag.Bool("useconf", false, "read HJSON/JSON config from stdin")
	useconffile := flag.String("useconffile", "", "read HJSON/JSON config from specified file path")
	normaliseconf := flag.Bool("normaliseconf", false, "use in combination with either -useconf or -useconffile, outputs your configuration normalised")
	exportkey := flag.Bool("exportkey", false, "use in combination with either -useconf or -useconffile, outputs your private key in PEM format")
	confjson := flag.Bool("json", false, "print configuration from -genconf or -normaliseconf as JSON instead of HJSON")
	autoconf := flag.Bool("autoconf", false, "automatic mode (dynamic IP, peer with IPv6 neighbors)")
	ver := flag.Bool("version", false, "prints the version of this build")
	logto := flag.String("logto", "stdout", "file path to log to, \"syslog\" or \"stdout\"")
	getaddr := flag.Bool("address", false, "use in combination with either -useconf or -useconffile, outputs your IPv6 address")
	getsnet := flag.Bool("subnet", false, "use in combination with either -useconf or -useconffile, outputs your IPv6 subnet")
	getpkey := flag.Bool("publickey", false, "use in combination with either -useconf or -useconffile, outputs your public key")
	loglevel := flag.String("loglevel", "info", "loglevel to enable")
	httpaddress := flag.String("httpaddress", "", "httpaddress to enable")
	wwwroot := flag.String("wwwroot", "", "wwwroot to enable")

	flag.Parse()

	// Create a new logger that logs output to stdout.
	var logger *log.Logger
	switch *logto {
	case "stdout":
		logger = log.New(os.Stdout, "", log.Flags())

	case "syslog":
		if syslogger, err := gsyslog.NewLogger(gsyslog.LOG_NOTICE, "DAEMON", version.BuildName()); err == nil {
			logger = log.New(syslogger, "", log.Flags()&^(log.Ldate|log.Ltime))
		}

	default:
		if logfd, err := os.OpenFile(*logto, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			logger = log.New(logfd, "", log.Flags())
		}
	}
	if logger == nil {
		logger = log.New(os.Stdout, "", log.Flags())
		logger.Warnln("Logging defaulting to stdout")
	}
	if *normaliseconf {
		setLogLevel("error", logger)
	} else {
		setLogLevel(*loglevel, logger)
	}

	cfg := config.GenerateConfig()
	var err error
	switch {
	case *ver:
		fmt.Println("Build name:", version.BuildName())
		fmt.Println("Build version:", version.BuildVersion())
		return
	case *autoconf:
		// Use an autoconf-generated config, this will give us random keys and
		// port numbers, and will use an automatically selected TUN interface.
	case *useconf:
		if _, err := cfg.ReadFrom(os.Stdin); err != nil {
			panic(err)
		}
	case *useconffile != "":
		f, err := os.Open(*useconffile)
		if err != nil {
			panic(err)
		}
		if _, err := cfg.ReadFrom(f); err != nil {
			panic(err)
		}
		_ = f.Close()

	case *genconf:
		var bs []byte
		if *confjson {
			bs, err = json.MarshalIndent(cfg, "", "  ")
		} else {
			bs, err = hjson.Marshal(cfg)
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
		return
	default:
		fmt.Println("Usage:")
		flag.PrintDefaults()

		if *getaddr || *getsnet {
			fmt.Println("\nError: You need to specify some config data using -useconf or -useconffile.")
		}
		return
	}

	privateKey := ed25519.PrivateKey(cfg.PrivateKey)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	n := &node{}

	getNodeKey := func() types.Domain {
		name := cfg.Domain
		return types.Domain{Key: publicKey, Name: []byte(name)}
	}

	switch {
	case *getaddr:
		if key := getNodeKey(); !key.Equal(types.Domain{}) {
			addr := n.core.AddrForDomain(key)
			ip := net.IP(addr[:])
			fmt.Println(ip.String())
		}
		return
	case *getsnet:
		if key := getNodeKey(); !key.Equal(types.Domain{}) {
			snet := n.core.SubnetForDomain(key)
			ipnet := net.IPNet{
				IP:   append(snet[:], 0, 0, 0, 0, 0, 0, 0, 0),
				Mask: net.CIDRMask(len(snet)*8, 128),
			}
			fmt.Println(ipnet.String())
		}
		return
	case *getpkey:
		fmt.Println(hex.EncodeToString(publicKey))
		return
	case *normaliseconf:
		if cfg.PrivateKeyPath != "" {
			cfg.PrivateKey = nil
		}
		var bs []byte
		if *confjson {
			bs, err = json.MarshalIndent(cfg, "", "  ")
		} else {
			bs, err = hjson.Marshal(cfg)
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bs))
		return
	case *exportkey:
		pem, err := cfg.MarshalPEMPrivateKey()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(pem))
		return
	}
	// Have we got a working configuration? If we don't then it probably means
	// that neither -autoconf, -useconf or -useconffile were set above.
	if cfg == nil {
		return
	}

	// Setup the RiV-mesh node itself.
	{
		options := []core.SetupOption{
			core.Domain(cfg.Domain),
			core.NodeInfo(cfg.NodeInfo),
			core.NodeInfoPrivacy(cfg.NodeInfoPrivacy),
			core.NetworkDomain(cfg.NetworkDomain),
			core.DDnsServer(cfg.DDnsServer),
		}
		for _, addr := range cfg.Listen {
			options = append(options, core.ListenAddress(addr))
		}
		for _, peer := range cfg.Peers {
			options = append(options, core.Peer{URI: peer})
		}
		for intf, peers := range cfg.InterfacePeers {
			for _, peer := range peers {
				options = append(options, core.Peer{URI: peer, SourceInterface: intf})
			}
		}
		for _, allowed := range cfg.AllowedPublicKeys {
			k, err := hex.DecodeString(allowed)
			if err != nil {
				panic(err)
			}
			options = append(options, core.AllowedPublicKey(k[:]))
		}
		if n.core, err = core.New(cfg.Certificate, logger, options...); err != nil {
			panic(err)
		}
	}

	// Setup the DNS.
	{
		if n.dns_server, err = dnsapi.NewDnsServer(cfg.Domain, dnsapi.DnsServerCfg{
			Core:            n.core,
			Tld:             cfg.DDnsServer.Tld,
			ListenAddress:   cfg.DDnsServer.ListenAddress,
			UpstreamServers: cfg.DDnsServer.UpstreamServers,
			Log:             logger,
		}); err != nil {
			logger.Errorln(err)
		} else {
			n.dns_server.Run()
		}
	}

	// Setup the multicast module.
	{
		options := []multicast.SetupOption{}
		for _, intf := range cfg.MulticastInterfaces {
			options = append(options, multicast.MulticastInterface{
				Regex:    regexp.MustCompile(intf.Regex),
				Beacon:   intf.Beacon,
				Listen:   intf.Listen,
				Port:     intf.Port,
				Priority: uint8(intf.Priority),
			})
		}
		if n.multicast, err = multicast.New(n.core, logger, options...); err != nil {
			fmt.Println("Multicast module fail:", err)
		}
	}

	// Setup the TUN module.
	{
		options := []tun.SetupOption{
			tun.InterfaceName(cfg.IfName),
			tun.InterfaceMTU(cfg.IfMTU),
		}
		if n.tun, err = tun.New(n.core, logger, options...); err != nil {
			panic(err)
		}
	}

	// Setup the REST socket.
	{
		//override httpaddress and wwwroot parameters in cfg
		if len(cfg.HttpAddress) == 0 {
			cfg.HttpAddress = *httpaddress
		}
		if len(cfg.WwwRoot) == 0 {
			cfg.WwwRoot = *wwwroot
		}
		cfg.HttpAddress = strings.Replace(cfg.HttpAddress, "<tun>", "["+n.core.Address().String()+"]", 1)

		if n.rest_server, err = restapi.NewRestServer(restapi.RestServerCfg{
			Core:          n.core,
			Multicast:     n.multicast,
			Log:           logger,
			ListenAddress: cfg.HttpAddress,
			WwwRoot:       cfg.WwwRoot,
			ConfigFn:      *useconffile,
			Features:      []string{},
		}); err != nil {
			logger.Errorln(err)
		} else {
			err = n.rest_server.Serve()
			if err != nil {
				logger.Errorln(err)
			}
		}
	}

	// Make some nice output that tells us what our IPv6 address and subnet are.
	// This is just logged to stdout for the user.
	address := n.core.Address()
	subnet := n.core.Subnet()
	public := n.core.GetSelf().Domain
	logger.Infof("Your Domain is %s", string(public.GetNormalizedName())+n.core.GetDdnsServer().Tld)
	logger.Infof("Your public key is %s", hex.EncodeToString(public.Key[:]))
	logger.Infof("Your IPv6 address is %s", address.String())
	logger.Infof("Your IPv6 subnet is %s", subnet.String())

	//Windows service shutdown service
	minwinsvc.SetOnExit(func() {
		logger.Infof("Shutting down service ...")
		sigCh <- os.Interrupt
		//there is a pause in handler. If the handler is finished other routines are not running.
		//Slee code gives a chance to run Stop methods.
		time.Sleep(10 * time.Second)
	})
	// Block until we are told to shut down.
	<-sigCh
	_ = n.multicast.Stop()
	_ = n.tun.Stop()
	n.core.Stop()
	err = n.rest_server.Shutdown()
	if err != nil {
		logger.Errorf("REST server shutdown error: %v", err)
	}
	err = n.dns_server.Shutdown()
	if err != nil {
		logger.Errorf("DNS server shutdown error: %v", err)
	}
}

func main() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		run(sigCh)
	}()
	wg.Wait()
}
