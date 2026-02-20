//go:build linux

package collector

import (
	"bufio"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/types"
)

type linuxCollector struct{}

func NewCollector() Collector {
	return &linuxCollector{}
}

func (c *linuxCollector) Collect(ctx context.Context) (*types.NetworkProbeResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := &types.NetworkProbeResult{
		Platform:  types.PlatformLinux,
		Timestamp: now,
		Logs:      []string{},
	}

	wifi, wifiLogs := collectLinuxWifi(ctx)
	res.Logs = append(res.Logs, wifiLogs...)
	if wifi != nil {
		res.Wifi = wifi
	}

	net, netLogs := collectLinuxNetwork(ctx)
	res.Logs = append(res.Logs, netLogs...)
	if net != nil {
		res.Network = net
	}

	return res, nil
}

func collectLinuxWifi(ctx context.Context) (*types.WifiNetworkInfo, []string) {
	logs := []string{}
	cmd := exec.CommandContext(ctx, "nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL,FREQ", "dev", "wifi")
	out, err := cmd.Output()
	if err != nil {
		logs = append(logs, "nmcli wifi failed")
		return nil, logs
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		if parts[0] != "yes" {
			continue
		}
		ssid := parts[1]
		signal, _ := strconv.Atoi(parts[2])
		freq, _ := strconv.Atoi(parts[3])
		band := types.Band24
		if freq > 4000 {
			band = types.Band5
		}
		signalDbm := (signal / 2) - 100
		wifi := &types.WifiNetworkInfo{
			SSID:          ssid,
			SignalDbm:     signalDbm,
			SignalQuality: signal,
			Bands: []types.WifiBandSpeed{
				{Band: band, LinkSpeedMbps: 0},
			},
		}
		logs = append(logs, "Linux wifi collected")
		return wifi, logs
	}
	return nil, logs
}

func collectLinuxNetwork(ctx context.Context) (*types.NetworkDetails, []string) {
	logs := []string{}
	cmd := exec.CommandContext(ctx, "ip", "addr", "show")
	out, err := cmd.Output()
	if err != nil {
		logs = append(logs, "ip addr show failed")
		return nil, logs
	}
	localIP := ""
	subnet := ""
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "inet ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ipCidr := parts[1]
				ipParts := strings.Split(ipCidr, "/")
				if len(ipParts) == 2 {
					localIP = ipParts[0]
					if plen, e := strconv.Atoi(ipParts[1]); e == nil {
						subnet = prefixToMask(plen)
					}
					break
				}
			}
		}
	}
	cmdRoute := exec.CommandContext(ctx, "ip", "route", "show", "default")
	outRoute, err := cmdRoute.Output()
	gateway := ""
	if err == nil {
		line := strings.TrimSpace(strings.Split(string(outRoute), "\n")[0])
		parts := strings.Fields(line)
		for i := 0; i < len(parts); i++ {
			if parts[i] == "via" && i+1 < len(parts) {
				gateway = parts[i+1]
				break
			}
		}
	} else {
		logs = append(logs, "ip route show default failed")
	}
	dns := []string{}
	cmdDNS := exec.CommandContext(ctx, "bash", "-c", "grep nameserver /etc/resolv.conf | awk '{print $2}'")
	outDNS, err := cmdDNS.Output()
	if err == nil {
		for _, line := range strings.Split(string(outDNS), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				dns = append(dns, line)
			}
		}
	} else {
		logs = append(logs, "resolv.conf parse failed")
	}
	if localIP == "" && gateway == "" {
		return nil, logs
	}
	net := &types.NetworkDetails{
		LocalIP:    localIP,
		Gateway:    gateway,
		DNS:        dns,
		SubnetMask: subnet,
	}
	logs = append(logs, "Linux network collected")
	return net, logs
}
