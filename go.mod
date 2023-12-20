module github.com/RiV-chain/RiV-mesh

go 1.21

replace github.com/Arceliar/ironwood => github.com/RiV-chain/ironwood v0.0.0-20231220110601-a30286562e6e

replace github.com/mikispag/dns-over-tls-forwarder => github.com/RiV-chain/dns-over-tls-forwarder v0.0.0-20230828114909-c2cd9f8d79d3

require (
	github.com/Arceliar/phony v0.0.0-20220903101357-530938a4b13d
	github.com/apernet/quic-go v0.40.1-0.20231112225043-e7f3af208dee
	github.com/gologme/log v1.3.0
	github.com/hashicorp/go-syslog v1.0.0
	github.com/hjson/hjson-go v3.1.0+incompatible
	github.com/kardianos/minwinsvc v1.0.2
	github.com/miekg/dns v1.1.55
	github.com/mitchellh/mapstructure v1.4.1
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.11.0
	golang.org/x/text v0.12.0
	golang.zx2c4.com/wireguard v0.0.0-20211017052713-f87e87af0d9a
	golang.zx2c4.com/wireguard/windows v0.5.3
)

require gerace.dev/zipfs v0.2.0

require (
	github.com/Arceliar/ironwood v0.0.0-20231127131626-465b82dfb5bd
	github.com/eknkc/basex v1.0.1
	github.com/mikispag/dns-over-tls-forwarder v0.0.0-20230401080233-dae75d4680fd
	github.com/slonm/tableprinter v0.0.0-20230107100804-643098716018
	github.com/vorot93/golang-signals v0.0.0-20170221070717-d9e83421ce45
	golang.org/x/crypto v0.12.0
	golang.org/x/exp v0.0.0-20221217163422-3c43f8badb15
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224
)

require github.com/kataras/tablewriter v0.0.0-20180708051242-e063d29b7c23 // indirect

require (
	github.com/bits-and-blooms/bitset v1.5.0 // indirect
	github.com/bits-and-blooms/bloom/v3 v3.3.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/quic-go/qtls-go1-20 v0.4.1 // indirect
	github.com/rivo/uniseg v0.3.4 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	go.uber.org/mock v0.3.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
)

require (
	github.com/ip2location/ip2location-go/v9 v9.5.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
)
