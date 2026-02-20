package types

type Band string

const (
	Band24 Band = "2_4GHz"
	Band5  Band = "5GHz"
)

type WifiBandSpeed struct {
	Band             Band   `json:"band"`
	LinkSpeedMbps    int    `json:"linkSpeedMbps"`
	MeasuredSpeedMbps int   `json:"measuredSpeedMbps,omitempty"`
}

type WifiNetworkInfo struct {
	SSID          string          `json:"ssid"`
	BSSID         string          `json:"bssid,omitempty"`
	SignalDbm     int             `json:"signalDbm,omitempty"`
	SignalQuality int             `json:"signalQuality,omitempty"`
	Bands         []WifiBandSpeed `json:"bands"`
}

type NetworkDetails struct {
	LocalIP    string   `json:"localIp"`
	Gateway    string   `json:"gateway"`
	DNS        []string `json:"dns"`
	SubnetMask string   `json:"subnetMask"`
}

type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformLinux   Platform = "linux"
	PlatformMacOS   Platform = "macos"
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

type NetworkProbeResult struct {
	Platform   Platform         `json:"platform"`
	Timestamp  string           `json:"timestamp"`
	Wifi       *WifiNetworkInfo `json:"wifi,omitempty"`
	Network    *NetworkDetails  `json:"network,omitempty"`
	Errors     []string         `json:"errors,omitempty"`
	Logs       []string         `json:"logs,omitempty"`
	DurationMs int64            `json:"durationMs"`
}
