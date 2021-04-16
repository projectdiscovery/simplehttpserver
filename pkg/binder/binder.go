package binder

import (
	"fmt"
	"net"

	"github.com/phayes/freeport"
	"github.com/projectdiscovery/gologger"
)

// CanListenOn the specified address
func CanListenOn(address string) bool {
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		return false
	}
	if err := listener.Close(); err != nil {
		gologger.Info().Msgf("%s\n", err)
	}
	return true
}

// GetRandomListenAddress from the specified one
func GetRandomListenAddress(currentAddress string) (string, error) {
	addrOrig, _, err := net.SplitHostPort(currentAddress)
	if err != nil {
		return "", err
	}

	newPort, err := freeport.GetFreePort()
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(addrOrig, fmt.Sprintf("%d", newPort)), nil
}
