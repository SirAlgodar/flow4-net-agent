//go:build windows

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

type windowsCollector struct{}

func NewCollector() Collector {
	return &windowsCollector{}
}

func (c *windowsCollector) Collect(ctx context.Context) (*types.NetworkProbeResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := &types.NetworkProbeResult{
		Platform:  types.PlatformWindows,
		Timestamp: now,
		Logs:      []string{},
	}

	wifi, wifiLogs := collectWindowsWifi(ctx)
	res.Logs = append(res.Logs, wifiLogs...)
	if wifi != nil {
		res.Wifi = wifi
	}

	net, netLogs := collectWindowsNetwork(ctx)
	res.Logs = append(res.Logs, netLogs...)
	if net != nil {
		res.Network = net
	}

	return res, nil
}

func collectWindowsWifi(ctx context.Context) (*types.WifiNetworkInfo, []string) {
	logs := []string{}
	cmd := exec.CommandContext(ctx, "netsh", "wlan", "show", "interfaces")
	out, err := cmd.Output()
	if err != nil {
		logs = append(logs, "netsh wlan show interfaces failed")
		return nil, logs
	}
	ssid := ""
	signal := 0
	bssid := ""
	freqBand := types.Band24

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "SSID") && !strings.Contains(line, "BSSID") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ssid = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "BSSID") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				bssid = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "Signal") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(strings.TrimSuffix(parts[1], "%"))
				if v, e := strconv.Atoi(val); e == nil {
					signal = v
				}
			}
		}
		if strings.HasPrefix(line, "Radio type") {
			if strings.Contains(line, "802.11ac") {
				freqBand = types.Band5
			}
		}
	}
	if ssid == "" {
		return nil, logs
	}
	signalDbm := (signal / 2) - 100
	wifi := &types.WifiNetworkInfo{
		SSID:          ssid,
		BSSID:         bssid,
		SignalDbm:     signalDbm,
		SignalQuality: signal,
		Bands: []types.WifiBandSpeed{
			{Band: freqBand, LinkSpeedMbps: 0},
		},
	}
	logs = append(logs, "Windows wifi collected")
	return wifi, logs
}

func collectWindowsNetwork(ctx context.Context) (*types.NetworkDetails, []string) {
	logs := []string{}
	cmd := exec.CommandContext(ctx, "powershell", "-Command", "Get-NetIPConfiguration | Where-Object {$_.IPv4DefaultGateway -ne $null} | Select-Object -First 1 | ConvertTo-Json")
	out, err := cmd.Output()
	if err != nil {
		logs = append(logs, "Get-NetIPConfiguration failed")
		return nil, logs
	}
	localIP := ""
	gateway := ""
	subnet := ""
	dns := []string{}

	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if strings.Contains(line, "\"IPv4Address\"") && strings.Contains(line, "IpAddress") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				localIP = strings.Trim(strings.TrimSpace(parts[1]), "\",")
			}
		}
		if strings.Contains(line, "\"IPv4DefaultGateway\"") && strings.Contains(line, "NextHop") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				gateway = strings.Trim(strings.TrimSpace(parts[1]), "\",")
			}
		}
		if strings.Contains(line, "\"PrefixLength\"") && subnet == "" {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				if plen, e := strconv.Atoi(strings.TrimSpace(parts[1])); e == nil {
					subnet = prefixToMask(plen)
				}
			}
		}
	}

	cmdDNS := exec.CommandContext(ctx, "powershell", "-Command", "(Get-DnsClientServerAddress -AddressFamily IPv4 | Select-Object -ExpandProperty ServerAddresses | ConvertTo-Json)")
	outDNS, err := cmdDNS.Output()
	if err == nil {
		for _, line := range strings.Split(string(outDNS), "\n") {
			line = strings.Trim(line, " \"[],")
			if line != "" {
				dns = append(dns, line)
			}
		}
	} else {
		logs = append(logs, "Get-DnsClientServerAddress failed")
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
	logs = append(logs, "Windows network collected")
	return net, logs
}

func prefixToMask(prefix int) string {
	if prefix < 0 || prefix > 32 {
		return ""
	}
	maskParts := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		bits := 8
		if prefix < 8 {
			bits = prefix
		}
		val := 0
		for j := 0; j < bits; j++ {
			val = val | (1 << (7 - j))
		}
		maskParts = append(maskParts, strconv.Itoa(val))
		prefix -= bits
		if prefix < 0 {
			prefix = 0
		}
	}
	return strings.Join(maskParts, ".")
}
