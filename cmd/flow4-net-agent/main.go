package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/SirAlgodar/flow4-net-agent/pkg/collector"
)

func main() {
	jsonOut := flag.Bool("json", false, "output JSON")
	timeoutMs := flag.Int("timeout-ms", 5000, "timeout in milliseconds")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "ping-agent":
			fmt.Println("ok")
			return
		case "version":
			fmt.Println("flow4-net-agent v0.1.0")
			return
		}
	}

	c := collector.NewCollector()
	res := collector.CollectWithTimeout(c, time.Duration(*timeoutMs)*time.Millisecond)

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res.Probe)
		return
	}

	if res.Err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", res.Err)
	}
	b, err := json.MarshalIndent(res.Probe, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal error: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(b)
}
