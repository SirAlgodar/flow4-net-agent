//go:build android

package collector

import (
	"context"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/types"
)

type androidCollector struct{}

func NewCollector() Collector {
	return &androidCollector{}
}

func (c *androidCollector) Collect(ctx context.Context) (*types.NetworkProbeResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := &types.NetworkProbeResult{
		Platform:  types.PlatformAndroid,
		Timestamp: now,
		Logs:      []string{"android collector stub"},
		Errors:    []string{"native Android integration not implemented"},
	}
	return res, nil
}
