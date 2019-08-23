package main

import (
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"net"
)

var (
	address *string
	udp     *bool
)

func init() {
	common.Init("canconnect", "1.0.0", "2019", "Can connect to server:port", "mpetavy", common.APACHE, "https://github.com/mpetavy/canconnect", false, nil, nil, run, 0)

	address = flag.String("c", "", "server:port to test")
}

func run() error {
	conn, err := net.Dial("tcp", *address)
	if err != nil {
		return err
	}

	defer conn.Close()

	fmt.Printf("%s connected successfully ", *address)

	return nil
}

func main() {
	defer common.Cleanup()

	common.Run([]string{"c"})
}
