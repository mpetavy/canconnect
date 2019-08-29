package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mpetavy/common"
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
	p := strings.Index(*address, "/")

	if p == -1 {
		conn, err := net.Dial("tcp", *address)
		if err != nil {
			return err
		}

		defer conn.Close()

		fmt.Printf("%s connected successfully ", *address)

		return nil
	}

	var ip net.IP
	var ipNet *net.IPNet

	ip, ipNet, err := net.ParseCIDR(*address)
	if err != nil {
		panic(err)
	}

	ip = ip.To4()

	ones, bits := ipNet.Mask.Size()
	mask := net.CIDRMask(ones, bits)

	lastIp := net.IP(make([]byte, 4))
	for i := range ip {
		lastIp[i] = ip[i] | ^mask[i]
	}

	if lastIp[3] == 255 {
		lastIp[3]--
	}

	successIps := make(chan string, lastIp[3])
	wg := sync.WaitGroup{}

	var i byte

	for i = 1; i <= lastIp[3]; i++ {
		pingIp := fmt.Sprintf("%d.%d.%d.%d", lastIp[0], lastIp[1], lastIp[2], i)

		wg.Add(1)

		go func(pingIp string) {
			defer func() {
				common.Debug("Ping %s ended\n", pingIp)
				wg.Done()
			}()

			common.Debug("Ping %s ...\n", pingIp)

			cmd := exec.Command("ping.exe", "-n", "1", pingIp)

			err := common.Watchdog(cmd, time.Second)
			if err != nil {
				if _, ok := err.(*exec.ExitError); ok {
					return
				}

				if _, ok := err.(*common.ErrWatchdog); ok {
					return
				}
			}

			successIps <- pingIp
		}(pingIp)
	}

	wg.Wait()

	close(successIps)

	for ip := range successIps {
		common.Info("Ping %s successfull\n", ip)
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run([]string{"c"})
}
