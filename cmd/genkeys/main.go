/*
This file generates crypto keys.
It prints out a new set of keys each time if finds a "better" one.
By default, "better" means a higher NodeID (-> higher IP address).
This is because the IP address format can compress leading 1s in the address, to increase the number of ID bits in the address.

If run with the "-sig" flag, it generates signing keys instead.
A "better" signing key means one with a higher TreeID.
This only matters if it's high enough to make you the root of the tree.
*/
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/gologme/log"

	c "github.com/RiV-chain/RiV-mesh/src/core"
)

type keySet struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func main() {
	threads := runtime.GOMAXPROCS(0)
	var currentBest ed25519.PublicKey
	newKeys := make(chan keySet, threads)
	for i := 0; i < threads; i++ {
		go doKeys(newKeys)
	}
	for {
		newKey := <-newKeys
		if isBetter(currentBest, newKey.pub) || len(currentBest) == 0 {
			currentBest = newKey.pub
			fmt.Println("-----")
			fmt.Println("Priv:", hex.EncodeToString(newKey.priv))
			fmt.Println("Pub:", hex.EncodeToString(newKey.pub))
			logger := log.New(os.Stdout, "", log.Flags())
			core, _ := c.New(newKey.priv, logger, nil)
			addr := core.AddrForKey(newKey.pub)
			fmt.Println("IP:", net.IP(addr[:]).String())
		}
	}
}

func isBetter(oldPub, newPub ed25519.PublicKey) bool {
	for idx := range oldPub {
		if newPub[idx] < oldPub[idx] {
			return true
		}
		if newPub[idx] > oldPub[idx] {
			break
		}
	}
	return false
}

func doKeys(out chan<- keySet) {
	bestKey := make(ed25519.PublicKey, ed25519.PublicKeySize)
	for idx := range bestKey {
		bestKey[idx] = 0xff
	}
	for {
		pub, priv, err := ed25519.GenerateKey(nil)
		if err != nil {
			panic(err)
		}
		if !isBetter(bestKey, pub) {
			continue
		}
		bestKey = pub
		out <- keySet{priv, pub}
	}
}
