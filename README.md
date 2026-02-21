# RiV-mesh - Decentralized IPv6 Mesh Network

## Introduction

RiV-mesh is the legacy implementation of a fully end-to-end encrypted IPv6
mesh network, designed to provide secure connectivity between a wide spectrum of endpoint devices like IoT devices,
desktop computers or even routers.
It is lightweight, self-arranging, supported on multiple
platforms and allows pretty much any IPv6-capable application
to communicate securely with other network nodes.

## Supported Platforms

RiV-mesh works on a number of platforms, including Linux, macOS, Ubiquiti
EdgeRouter, VyOS, Windows, FreeBSD, OpenBSD and OpenWrt.

## Building

If you want to build from source, as opposed to installing one of the pre-built
packages:

1. Install [Go](https://golang.org) (requires Go 1.19 or later)
2. Clone this repository
2. Run `./build`

Note that you can cross-compile for other platforms and architectures by
specifying the `GOOS` and `GOARCH` environment variables, e.g. `GOOS=windows
./build` or `GOOS=linux GOARCH=mipsle ./build`

... or generate an iOS framework with:

```
./contrib/mobile/build -i
```

... or generate an Android AAR bundle with:

```
./contrib/mobile/build -a
```

## Running

### Generate configuration

To generate static configuration, either generate a HJSON file (human-friendly,
complete with comments):

```
./mesh -genconf > /path/to/mesh.conf
```

... or generate a plain JSON file (which is easy to manipulate
programmatically):

```
./mesh -genconf -json > /path/to/mesh.conf
```

You will need to edit the `mesh.conf` file to add or remove peers, modify
other configuration such as listen addresses or multicast addresses, etc.

### Run RiV-mesh

To run with the generated static configuration:

```
./mesh -useconffile /path/to/mesh.conf
```

To run in auto-configuration mode (which will use sane defaults and random keys
at each startup, instead of using a static configuration file):

```
./mesh -autoconf
```

You will likely need to run RiV-mesh as a privileged user or under `sudo`,
unless you have permission to create TUN/TAP adapters. On Linux this can be done
by giving the RiV-mesh binary the `CAP_NET_ADMIN` capability.

## Known issues

### 1. Log message:
```
An error occurred starting multicast: listen udp6 [::]:9001: socket: address family not supported by protocol
```
and
```
An error occurred starting TUN/TAP: operation not supported
```

### Caused by:
The device has no IPv6 support


### 2. Log message:
```
An error occurred starting TUN/TAP: permission denied
```

### Caused by:
IPv6 support is not enabled. See the solution: https://github.com/yggdrasil-network/yggdrasil-go/issues/479#issuecomment-519512395

### 3. Mesh infinite output in log:
 Connected SCTP ...
 
 Disconnected SCTP ...

### Caused by:
Docker interface docker0 is conflicting with SCTP bind process. The issue can be resolved by removing docker.

## License

RiV-mesh is released under the **GNU Lesser General Public License v3.0 (LGPLv3)**:

- **License**: LGPLv3 (standard version, no special exceptions)
- **Commercial Use**: Allowed with full LGPLv3 compliance requirements
- **Application Code Sharing**: Must share source code of applications using RiV-mesh
- **Build Instructions**: Must provide installation and build information
- **Modifications**: Any modifications to RiV-mesh must remain open source
