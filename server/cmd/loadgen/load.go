package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	insecurecreds "google.golang.org/grpc/credentials/insecure"
	"toolplane/internal/server"
	proto "toolplane/proto"
)

const (
	// #nosec G101 -- development fixture key for the embedded server this
	// tool boots itself; not a credential in any deployed sense.
	embeddedAPIKey = "toolplane-loadgen-key"
	echoTool       = "load.echo"
	streamTool     = "load.stream"
	// Provider claim poll cadence when idle; also the tick that turns a
	// session's queue latency into claim latency.
	idlePollInterval = 50 * time.Millisecond
)

type harness struct {
	cfg   config
	conn  *grpc.ClientConn
	tools proto.ToolServiceClient
	sess  proto.SessionsServiceClient
	mach  proto.MachinesServiceClient
	reqs  proto.RequestsServiceClient

	collector  *collector
	sessions   []string
	machines   []machineRef
	stop       chan struct{}
	stopOnce   sync.Once
	deadlineAt time.Time
	workerWG   sync.WaitGroup // shape workers — drain before providers stop
	wg         sync.WaitGroup // fake providers
}

// shutdown stops the fake providers exactly once and waits for them, so
// the report never misses in-flight provider work. Shape workers must
// have drained first (runShape does that) — stopping providers earlier
// would turn every in-flight consumer wait into a timeout.
func (h *harness) shutdown() {
	h.stopOnce.Do(func() { close(h.stop) })
	h.wg.Wait()
}

// pastDeadline reports whether the load window has closed; ops that fail
// because the window ended are recorded as partial, not as failures.
func (h *harness) pastDeadline() bool {
	return time.Now().After(h.deadlineAt)
}

type machineRef struct {
	sessionID string
	machineID string
}

func runLoad(ctx context.Context, cfg config) int {
	var (
		addr       string
		metricsURL = cfg.metricsURL
		shutdown   func()
	)

	if cfg.mode == "embedded" {
		addr, metricsURL, shutdown = bootEmbeddedServer()
		defer shutdown()
	} else {
		addr = cfg.address
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecurecreds.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: dial %s: %v\n", addr, err)
		return 1
	}
	defer conn.Close()

	h := &harness{
		cfg:       cfg,
		conn:      conn,
		tools:     proto.NewToolServiceClient(conn),
		sess:      proto.NewSessionsServiceClient(conn),
		mach:      proto.NewMachinesServiceClient(conn),
		reqs:      proto.NewRequestsServiceClient(conn),
		collector: newCollector(),
		stop:      make(chan struct{}),
	}

	callCtx, cancel := context.WithTimeout(ctx, cfg.duration+90*time.Second)
	defer cancel()
	callCtx = withAPIKey(callCtx, cfg.apiKey)

	if err := h.waitReady(callCtx); err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: server never became ready: %v\n", err)
		return 1
	}
	if err := h.setupSessions(callCtx); err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: setup: %v\n", err)
		h.shutdown()
		return 1
	}

	h.deadlineAt = time.Now().Add(cfg.duration)
	startMetrics := scrapeMetrics(metricsURL)
	runShape(callCtx, h)
	// Drain the providers after the consumers so their resolve/append ops
	// land in the counts, then scrape with every counter settled (the
	// server itself stays up until runLoad returns).
	h.shutdown()
	endMetrics := scrapeMetrics(metricsURL)

	report := h.collector.report(cfg, metricsDelta(startMetrics, endMetrics))
	encoded, _ := json.MarshalIndent(report, "", "  ")
	if cfg.reportPath != "" {
		if err := os.WriteFile(cfg.reportPath, append(encoded, '\n'), 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "loadgen: write report: %v\n", err)
		}
	}
	fmt.Println(string(encoded))
	if report.Failed {
		return 1
	}
	return 0
}

// bootEmbeddedServer starts an in-process control plane on a loopback
// listener and returns its address, metrics URL, and a shutdown func.
// Ambient TOOLPLANE_DATABASE_URL keeps working: set it to load-test the
// Postgres store without changing anything here.
func bootEmbeddedServer() (addr, metricsURL string, shutdown func()) {
	os.Setenv("TOOLPLANE_ENV_MODE", "development")
	os.Setenv("TOOLPLANE_AUTH_MODE", "fixed")
	os.Setenv("TOOLPLANE_AUTH_FIXED_API_KEY", embeddedAPIKey)
	if strings.TrimSpace(os.Getenv("TOOLPLANE_DATABASE_URL")) == "" &&
		strings.TrimSpace(os.Getenv("TOOLPLANE_STORAGE_MODE")) == "" {
		os.Setenv("TOOLPLANE_STORAGE_MODE", "memory")
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: listen: %v\n", err)
		os.Exit(1)
	}
	port := lis.Addr().(*net.TCPAddr).Port

	// Pick a metrics port by binding and releasing it — RunContext binds
	// the metrics listener itself, so a pre-bound listener here would
	// double-bind and the server would exit at boot.
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: metrics port probe: %v\n", err)
		os.Exit(1)
	}
	metricsAddr := probe.Addr().String()
	if err := probe.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "loadgen: metrics port probe close: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	opts := server.DefaultOptions()
	opts.Listener = lis
	opts.Port = port
	opts.MetricsListen = metricsAddr
	done := make(chan int, 1)
	go func() { done <- server.RunContext(ctx, opts) }()

	// Consume the exit so cancellation on shutdown is a clean return.
	go func() { <-done }()

	return fmt.Sprintf("127.0.0.1:%d", port),
		"http://" + metricsAddr,
		cancel
}

func (h *harness) waitReady(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		_, err := h.tools.HealthCheck(withAPIKey(pingCtx, h.cfg.apiKey), &proto.HealthCheckRequest{})
		cancel()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (h *harness) setupSessions(ctx context.Context) error {
	// Unique per invocation, not per process: against a persistent store
	// (Postgres, or any external server) a second run from the same
	// process must not collide with the first run's session rows.
	runID := time.Now().UnixNano()
	for i := 0; i < h.cfg.sessions; i++ {
		sessionID := fmt.Sprintf("loadgen-%d-%d", runID, i)
		if _, err := h.sess.CreateSession(withAPIKey(ctx, h.cfg.apiKey), &proto.CreateSessionRequest{
			UserId:      "loadgen",
			Name:        sessionID,
			Description: "loadgen session",
			SessionId:   sessionID,
		}); err != nil {
			return fmt.Errorf("create session %s: %w", sessionID, err)
		}

		machineID := "machine-" + sessionID
		if _, err := h.mach.RegisterMachine(withAPIKey(ctx, h.cfg.apiKey), &proto.RegisterMachineRequest{
			SessionId:   sessionID,
			MachineId:   machineID,
			SdkVersion:  "loadgen",
			SdkLanguage: "go",
			Tools: []*proto.RegisterToolRequest{
				{Name: echoTool, Description: "resolves immediately", Schema: `{"type":"object"}`},
				{Name: streamTool, Description: "streams chunks then resolves", Schema: `{"type":"object"}`},
			},
		}); err != nil {
			return fmt.Errorf("register machine %s: %w", machineID, err)
		}

		h.sessions = append(h.sessions, sessionID)
		h.machines = append(h.machines, machineRef{sessionID: sessionID, machineID: machineID})

		h.wg.Add(1)
		go h.runFakeProvider(sessionID, machineID)
	}
	return nil
}

// percentile of sorted sample durations, nearest-rank.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(float64(len(sorted)-1)*p + 0.5)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}
