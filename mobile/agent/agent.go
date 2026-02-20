package agent

import (
	"encoding/json"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/collector"
)

func CollectJson(timeoutMs int) string {
	c := collector.NewCollector()
	res := collector.CollectWithTimeout(c, time.Duration(timeoutMs)*time.Millisecond)
	if res.Probe == nil {
		return "{}"
	}
	b, err := json.Marshal(res.Probe)
	if err != nil {
		return "{}"
	}
	return string(b)
}
