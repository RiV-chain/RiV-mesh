//go:build linux || netbsd || freebsd || openbsd || dragonflybsd
// +build linux netbsd freebsd openbsd dragonflybsd

package multicast

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func (m *Multicast) _multicastStarted() {

}

func (m *Multicast) multicastReuse(network string, address string, c syscall.RawConn) error {
	var control error

	control = c.Control(func(fd uintptr) {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set SO_REUSEPORT on socket: %s\n", err)
		}
	})
	return control
}
