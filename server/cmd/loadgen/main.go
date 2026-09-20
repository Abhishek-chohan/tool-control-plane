// Command loadgen drives the Toolplane control plane with agent-shaped
// load and reports per-op latency, error counts, and server metric
// deltas. It exists to answer capacity questions against real hardware;
// its numbers are relative to wherever it runs.
//
// Modes:
//
//	embedded (default) — boots an in-process server (memory store, or
//	  Postgres when TOOLPLANE_DATABASE_URL is set in the environment)
//	  and drives it over loopback gRPC.
//	external — drives --address; storage/auth are whatever that server
//	  runs.
//
// Shapes: agent-turn, token-stream, heartbeat-fleet, discovery-churn,
// mixed (default). See shapes.go.
//
// Exit code: 0 when every driven op succeeded, 1 otherwise — so CI can
// treat a run with unexpected failures as red.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

type config struct {
	mode        string
	address     string
	apiKey      string
	shape       string
	drill       string
	sessions    int
	duration    time.Duration
	reportPath  string
	metricsURL  string
	streamCount int
	burst       int
}

func parseFlags(argv []string) (config, error) {
	fs := flag.NewFlagSet("loadgen", flag.ContinueOnError)
	cfg := config{}
	fs.StringVar(&cfg.mode, "mode", "embedded", "embedded boots an in-process server; external drives --address")
	fs.StringVar(&cfg.address, "address", "localhost:9001", "server address (external mode)")
	fs.StringVar(&cfg.apiKey, "api-key", "", "API key; defaults to $TOOLPLANE_API_KEY, then the embedded key")
	fs.StringVar(&cfg.shape, "shape", "mixed", "workload shape: agent-turn, token-stream, heartbeat-fleet, discovery-churn, mixed")
	fs.StringVar(&cfg.drill, "drill", "", "run a reliability drill under load instead of a plain shape: provider-kill, drain-under-backlog, multi-instance-contention")
	fs.IntVar(&cfg.sessions, "sessions", 8, "concurrent load sessions")
	fs.DurationVar(&cfg.duration, "duration", 60*time.Second, "drive load for this long")
	fs.StringVar(&cfg.reportPath, "report", "", "write the JSON report here (default: stdout only)")
	fs.StringVar(&cfg.metricsURL, "metrics-url", "", "server /metrics base URL; embedded mode discovers it automatically")
	fs.IntVar(&cfg.streamCount, "streams", 8, "concurrent token streams (token-stream and mixed shapes)")
	fs.IntVar(&cfg.burst, "burst", 10, "tool calls per agent turn (agent-turn and mixed shapes)")
	if err := fs.Parse(argv); err != nil {
		return config{}, err
	}
	if cfg.apiKey == "" {
		cfg.apiKey = os.Getenv("TOOLPLANE_API_KEY")
	}
	if cfg.apiKey == "" {
		cfg.apiKey = embeddedAPIKey
	}
	switch cfg.shape {
	case "agent-turn", "token-stream", "heartbeat-fleet", "discovery-churn", "mixed":
	default:
		return config{}, fmt.Errorf("unknown shape %q", cfg.shape)
	}
	switch cfg.drill {
	case "", "provider-kill", "drain-under-backlog", "multi-instance-contention":
	default:
		return config{}, fmt.Errorf("unknown drill %q", cfg.drill)
	}
	if cfg.duration <= 0 {
		return config{}, fmt.Errorf("--duration must be positive")
	}
	return cfg, nil
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: %v\n", err)
		os.Exit(2)
	}
	os.Exit(runLoad(context.Background(), cfg))
}
