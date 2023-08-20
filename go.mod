module github.com/RiV-chain/RiV-mesh

go 1.21

replace github.com/Arceliar/ironwood => github.com/RiV-chain/ironwood v0.0.0-20230820133607-9038e37c47b5

replace github.com/mikispag/dns-over-tls-forwarder => github.com/RiV-chain/dns-over-tls-forwarder v0.0.0-20230819192037-9ad070cc8a60

require (
	github.com/Arceliar/ironwood v0.0.0-20230805085300-86206813435f
	github.com/Arceliar/phony v0.0.0-20220903101357-530938a4b13d
	github.com/getlantern/multipath v0.0.0-20220920195041-55195f38df73
	github.com/gologme/log v1.3.0
	github.com/hashicorp/go-syslog v1.0.0
	github.com/hjson/hjson-go v3.1.0+incompatible
	github.com/kardianos/minwinsvc v1.0.2
	github.com/miekg/dns v1.1.55
	github.com/mitchellh/mapstructure v1.4.1
	github.com/vikulin/sctp v0.0.0-20221009200520-ae0f2830e422
	github.com/vishvananda/netlink v1.1.0
	golang.org/x/net v0.10.0
	golang.org/x/sys v0.11.0
	golang.org/x/text v0.12.0
	golang.zx2c4.com/wireguard v0.0.0-20211017052713-f87e87af0d9a
	golang.zx2c4.com/wireguard/windows v0.5.3
)

require gerace.dev/zipfs v0.2.0

require (
	github.com/mikispag/dns-over-tls-forwarder v0.0.0-20230401080233-dae75d4680fd
	github.com/slonm/tableprinter v0.0.0-20230107100804-643098716018
	github.com/vorot93/golang-signals v0.0.0-20170221070717-d9e83421ce45
	golang.org/x/exp v0.0.0-20221217163422-3c43f8badb15
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224
)

require github.com/kataras/tablewriter v0.0.0-20180708051242-e063d29b7c23 // indirect

require (
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/ema v0.0.0-20190620044903-5943d28f40e4 // indirect
	github.com/getlantern/errors v1.0.1 // indirect
	github.com/getlantern/golog v0.0.0-20211223150227-d4d95a44d873 // indirect
	github.com/getlantern/hex v0.0.0-20190417191902-c6586a6fe0b7 // indirect
	github.com/getlantern/hidden v0.0.0-20190325191715-f02dbb02be55 // indirect
	github.com/getlantern/ops v0.0.0-20190325191751-d70cb0d6f85f // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/libp2p/go-buffer-pool v0.0.2 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/rivo/uniseg v0.3.4 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
)

require (
	github.com/ip2location/ip2location-go/v9 v9.5.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
)
