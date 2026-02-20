//go:build ios

package collector

import (
	"context"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/types"
)

type iosCollector struct{}

func NewCollector() Collector {
	return &iosCollector{}
}

func (c *iosCollector) Collect(ctx context.Context) (*types.NetworkProbeResult, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res := &types.NetworkProbeResult{
		Platform:  types.PlatformIOS,
		Timestamp: now,
		Logs:      []string{"ios collector stub"},
		Errors:    []string{"native iOS integration not implemented"},
	}
	return res, nil
}
