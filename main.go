package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mpetavy/common"
)

var (
	address *string
)

func init() {
	common.Init("1.0.0", "2019", "Can connect to server:port", "mpetavy", common.APACHE, false, nil, nil, run, 0)

	address = flag.String("c", "", "server:port to test")
}

func run() error {
	port := -1

	p := strings.Index(*address, ":")

	if p != -1 {
		var err error

		port, err = strconv.Atoi((*address)[p+1:])
		if err != nil {
			return err
		}

		*address = (*address)[:p]
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
				common.Debug("%s ended\n", pingIp)
				wg.Done()
			}()

			common.Debug("%s ...\n", pingIp)

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

			pingIp = fmt.Sprintf("%s:%d", pingIp, port)

			if port != -1 {
				conn, err := net.Dial("tcp", pingIp)
				if err != nil {
					return
				}

				defer func() {
					common.WarnError(conn.Close())
				}()
			}

			successIps <- pingIp
		}(pingIp)
	}

	wg.Wait()

	close(successIps)

	for ip := range successIps {
		common.Info("%s\n", ip)
	}

	return nil
}

func main() {
	defer common.Done()

	common.Run([]string{"c"})
}
