package cli

import (
	"fmt"
	"net"
	"os"
	"time"
)

// Custom dialer for use with grpc dialer.
func adminServerDialer(address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, timeout)

	if err != nil {
		fmt.Printf("Could not connect - %s\n", err.Error())
		os.Exit(1)
	}

	return conn, err
}
