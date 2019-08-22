package main

import (
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"net"
	"time"
)

var (
	address *string
	udp     *bool
)

func init() {
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

	common.New(&common.App{"canconnect", "1.0.0", "2019", "Can connect to server:port", "mpetavy", common.APACHE, "https://github.com/mpetavy/canconnect", false, nil, nil, run, time.Duration(0)}, []string{"c"})
	common.Run()
}
