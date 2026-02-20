package collector

import (
	"context"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/types"
)

type Collector interface {
	Collect(ctx context.Context) (*types.NetworkProbeResult, error)
}

type Result struct {
	Probe *types.NetworkProbeResult
	Err   error
}

func CollectWithTimeout(base Collector, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ch := make(chan Result, 1)

	go func() {
		start := time.Now()
		probe, err := base.Collect(ctx)
		if probe != nil {
			probe.DurationMs = time.Since(start).Milliseconds()
		}
		ch <- Result{Probe: probe, Err: err}
	}()

	select {
	case r := <-ch:
		return r
	case <-ctx.Done():
		return Result{
			Probe: &types.NetworkProbeResult{
				Errors:     []string{"timeout exceeded"},
				DurationMs: timeout.Milliseconds(),
			},
			Err: ctx.Err(),
		}
	}
}
