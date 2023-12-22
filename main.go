package main

import (
	"crypto/tls"
	"embed"
	"flag"
	"fmt"
	"github.com/mpetavy/common"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	address *string
	useTls  *bool
)

//go:embed go.mod
var resources embed.FS

func init() {
	common.Init("", "", "", "", "Can connect to server:port", "", "", "", &resources, nil, nil, run, 0)

	address = flag.String("c", "", "server:port to test")
	useTls = flag.Bool("tls", false, "Use TLS")
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func process(ip string, port int, tlsConfig *tls.Config, successIps chan string) error {
	common.Debug("%s ...\n", ip)

	countFlag := "-n"
	if !common.IsWindows() {
		countFlag = "-c"
	}

	cmd := exec.Command("ping", countFlag, "1", ip)

	_, err := common.NewWatchdogCmd(cmd, time.Second*5)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return err
		}

		if _, ok := err.(*common.ErrWatchdog); ok {
			return err
		}
	}

	if port == 0 {
		successIps <- ip

		return nil
	}

	pingIp := fmt.Sprintf("%s:%d", ip, port)

	ep, connector, err := common.NewEndpoint(pingIp, true, tlsConfig)
	if common.Error(err) {
		return err
	}

	err = ep.Start()
	if common.Error(err) {
		return err
	}

	defer func() {
		common.Error(ep.Stop())
	}()

	connection, err := connector()
	if common.Error(err) {
		return err
	}

	defer func() {
		common.DebugError(connection.Close())
	}()

	successIps <- pingIp

	return nil
}

func run() error {
	var host string
	var port int
	var portstr string
	var err error
	var ips []string
	var tlsConfig *tls.Config

	if *useTls {
		tlsConfig, err = common.NewTlsConfigFromFlags()
		if common.Error(err) {
			return err
		}
	}

	if strings.HasPrefix(*address, ":") {
		*address = "localhost" + *address
	}

	host, portstr, err = net.SplitHostPort(*address)
	if err == nil {
		port, err = strconv.Atoi(portstr)
		if common.Error(err) {
			return err
		}

		ipaddr, err := net.ResolveIPAddr("", host)
		if common.Error(err) {
			return err
		}

		ips = append(ips, ipaddr.IP.String())
	} else {
		cip, cipnet, err := net.ParseCIDR(*address)
		if common.Error(err) {
			return err
		}
		for x := cip.Mask(cipnet.Mask); cipnet.Contains(x); inc(x) {
			ips = append(ips, x.String())
		}
	}

	wg := sync.WaitGroup{}
	successIps := make(chan string, len(ips))

	for _, ip := range ips {
		if strings.HasSuffix(ip, ".0") || strings.HasSuffix(ip, ".255") {
			continue
		}

		wg.Add(1)
		go func(ip string) {
			defer common.UnregisterGoRoutine(common.RegisterGoRoutine(1))

			defer wg.Done()

			common.Error(process(ip, port, tlsConfig, successIps))
		}(ip)
	}

	wg.Wait()

	close(successIps)

	for ip := range successIps {
		common.Info("%s\n", ip)
	}

	return nil
}

func main() {
	common.Run([]string{"c"})
}
