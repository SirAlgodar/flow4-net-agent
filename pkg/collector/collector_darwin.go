//go:build darwin

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

type darwinCollector struct{}

func NewCollector() Collector {
	return &darwinCollector{}
}

func (c *darwinCollector) Collect(ctx context.Context) (*types.NetworkProbeResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := &types.NetworkProbeResult{
		Platform:  types.PlatformMacOS,
		Timestamp: now,
		Logs:      []string{},
	}

	wifi, wifiLogs := collectDarwinWifi(ctx)
	res.Logs = append(res.Logs, wifiLogs...)
	if wifi != nil {
		res.Wifi = wifi
	}

	net, netLogs := collectDarwinNetwork(ctx)
	res.Logs = append(res.Logs, netLogs...)
	if net != nil {
		res.Network = net
	}

	return res, nil
}

func collectDarwinWifi(ctx context.Context) (*types.WifiNetworkInfo, []string) {
	logs := []string{}
	cmd := exec.CommandContext(ctx, "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport", "-I")
	out, err := cmd.Output()
	if err != nil {
		logs = append(logs, "airport -I failed")
		return nil, logs
	}
	ssid := ""
	rssi := 0
	freq := 0

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "SSID:") {
			ssid = strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		}
		if strings.HasPrefix(line, "agrCtlRSSI:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "agrCtlRSSI:"))
			rssi, _ = strconv.Atoi(val)
		}
		if strings.HasPrefix(line, "channel:") {
			fields := strings.Fields(strings.TrimPrefix(line, "channel:"))
			if len(fields) > 0 {
				freq, _ = strconv.Atoi(fields[0])
			}
		}
	}
	if ssid == "" {
		return nil, logs
	}
	band := types.Band24
	if freq > 30 {
		band = types.Band5
	}
	wifi := &types.WifiNetworkInfo{
		SSID:      ssid,
		SignalDbm: rssi,
		Bands: []types.WifiBandSpeed{
			{Band: band, LinkSpeedMbps: 0},
		},
	}
	logs = append(logs, "Darwin wifi collected")
	return wifi, logs
}

func collectDarwinNetwork(ctx context.Context) (*types.NetworkDetails, []string) {
	logs := []string{}
	cmdIP := exec.CommandContext(ctx, "ipconfig", "getifaddr", "en0")
	outIP, err := cmdIP.Output()
	localIP := ""
	if err == nil {
		localIP = strings.TrimSpace(string(outIP))
	} else {
		logs = append(logs, "ipconfig getifaddr en0 failed")
	}
	cmdGW := exec.CommandContext(ctx, "route", "-n", "get", "default")
	outGW, err := cmdGW.Output()
	gateway := ""
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(outGW)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "gateway:") {
				gateway = strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
				break
			}
		}
	} else {
		logs = append(logs, "route -n get default failed")
	}
	dns := []string{}
	cmdDNS := exec.CommandContext(ctx, "bash", "-c", "scutil --dns | awk '/nameserver\\[/{getline; print $1}'")
	outDNS, err := cmdDNS.Output()
	if err == nil {
		for _, line := range strings.Split(string(outDNS), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				dns = append(dns, line)
			}
		}
	} else {
		logs = append(logs, "scutil --dns failed")
	}
	if localIP == "" && gateway == "" {
		return nil, logs
	}
	net := &types.NetworkDetails{
		LocalIP:    localIP,
		Gateway:    gateway,
		DNS:        dns,
		SubnetMask: "",
	}
	logs = append(logs, "Darwin network collected")
	return net, logs
}
