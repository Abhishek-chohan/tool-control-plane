package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"toolplane/internal/cli"
	"toolplane/internal/server"
	"toolplane/pkg/service"
	proto "toolplane/proto"
)

// demoToolsFile is the tools module the demo serves: one add tool with
// the documented bare-@tool shape.
const demoToolsFile = `from toolplane.provider_registry import tool


@tool(name="add", description="Add two numbers")
def add(a: int, b: int) -> int:
    return a + b
`

// newDemoCommand runs the whole loop in one terminal: an in-memory
// control plane, a provider serving a scaffolded tools file, and an
// invoke — then tears everything down. --drill makes the durability
// story visible: the provider dies mid-request and the request still
// completes.
func newDemoCommand() *cobra.Command {
	var (
		drill        bool
		providerPath string
	)
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Run the full loop: server, provider, and invoke — in one terminal",
		Long: `Run the Toolplane loop end to end without touching anything else:

  1. boots an in-memory control plane (dev defaults, fixed dev key),
  2. starts a provider serving a tools file,
  3. invokes a tool and prints the result,
  4. tears everything down.

With --drill, the provider is killed while a request executes and the
request still completes — reclaimed, re-executed, and finished by the
control plane, which is the point of durable remote tools.

Requires the Python provider CLI (pip install toolplane-python-client).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return exitError{code: runDemo(cmd, providerPath, drill)}
		},
	}
	cmd.Flags().BoolVar(&drill, "drill", false, "kill the provider mid-request and show the request complete anyway")
	cmd.Flags().StringVar(&providerPath, "provider", "", "toolplane-provider binary to use (default: look up toolplane-provider on PATH)")
	return cmd
}

// runDemo orchestrates the loop. Everything is in-memory and local; the
// only external dependency is the Python provider CLI, which is looked
// up lazily with a teaching error when absent.
func runDemo(cmd *cobra.Command, providerPath string, drill bool) int {
	out := cmd.OutOrStdout()

	// The demo fixes its own environment contract: fixed dev key over
	// in-memory storage. Set explicitly so operator env vars can't flip
	// the demo into postgres mode or a failed config load.
	for _, kv := range [][2]string{
		{"TOOLPLANE_ENV_MODE", "development"},
		{"TOOLPLANE_AUTH_MODE", "fixed"},
		{"TOOLPLANE_AUTH_FIXED_API_KEY", "dev-key"},
		{"TOOLPLANE_STORAGE_MODE", "memory"},
	} {
		if err := os.Setenv(kv[0], kv[1]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return cli.ExitError
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Control plane on a kernel-assigned port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	port := lis.Addr().(*net.TCPAddr).Port
	opts := server.DefaultOptions()
	opts.Listener = lis
	opts.MetricsListen = ""
	// Build identity is set before the server goroutine starts: the
	// goroutine-creation edge orders this write before every serving
	// read (HealthCheck reports it). StorageSummary needs no assignment
	// here — the server resolves and publishes it from the storage mode.
	service.BuildVersion = version
	go func() { _ = server.RunContext(ctx, opts) }()

	// Readiness: the health flip happens after service registration.
	conn := cli.NewConnection(fmt.Sprintf("localhost:%d", port), "dev-key", 2*time.Second)
	readyCtx, readyCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer readyCancel()
	if err := cli.WaitReady(readyCtx, conn, 100*time.Millisecond); err != nil {
		fmt.Fprintf(os.Stderr, "control plane did not become ready: %v\n", err)
		return cli.ExitError
	}
	fmt.Fprintf(out, "==> control plane ready on %s (storage: memory, auth: fixed dev key)\n", conn.Address)

	// 2. Tools file: scaffold in a temp dir so the demo never touches cwd.
	dir, err := os.MkdirTemp("", "toolplane-demo-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer os.RemoveAll(dir)
	toolsFile := filepath.Join(dir, "tools.py")
	if err := os.WriteFile(toolsFile, []byte(demoToolsFile), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}

	// 3. Provider subprocess.
	providerBin := providerPath
	if providerBin == "" {
		providerBin = "toolplane-provider"
	}
	sessionID := fmt.Sprintf("demo-%d", time.Now().UnixNano())
	providerArgs := []string{"serve", toolsFile, "--session", sessionID,
		"--host", "localhost", "--port", fmt.Sprintf("%d", port), "--api-key", "dev-key"}
	providerCmd := exec.Command(providerBin, providerArgs...)
	// The subprocess writes straight to the terminal fds: fd passthrough
	// means no parent-side copier goroutine, so provider output streams
	// live AND nothing races the command's own writer (a shared buffer
	// here is a data race between the copier and this function's prints).
	providerCmd.Stdout = os.Stdout
	providerCmd.Stderr = os.Stderr
	if err := providerCmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\nthe provider CLI is missing — install it with:\n  pip install toolplane-python-client\n", err)
		return cli.ExitError
	}
	// The provider must not outlive the demo on any exit path. The happy
	// path returns before the drill's explicit kill, and an orphaned
	// provider also holds any inherited pipe open — a piped
	// `toolplane demo | tail` would never see the pipe close. Wait (not
	// Process.Wait) also joins the exec machinery.
	defer func() {
		_ = providerCmd.Process.Kill()
		_ = providerCmd.Wait()
	}()
	fmt.Fprintf(out, "==> provider serving tools.py in session %s\n", sessionID)

	// Wait for the provider to finish registering: poll the tool until it
	// resolves in the session (Python startup + registration take a beat).
	gconn, err := conn.Dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconn.Close()
	tools := proto.NewToolServiceClient(gconn)
	toolDeadline := time.Now().Add(20 * time.Second)
	for {
		toolCtx, toolCancel := context.WithTimeout(ctx, 2*time.Second)
		toolCtx = metadata.AppendToOutgoingContext(toolCtx, "api_key", "dev-key")
		_, err := tools.GetTool(toolCtx, &proto.GetToolRequest{
			SessionId: sessionID,
			ToolName:  "add",
		})
		toolCancel()
		if err == nil {
			break
		}
		if time.Now().After(toolDeadline) {
			fmt.Fprintln(os.Stderr, "provider never registered tool add — check the provider output above")
			return cli.ExitError
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !drill {
		callCtx, callCancel := context.WithTimeout(ctx, 30*time.Second)
		defer callCancel()
		callCtx = metadata.AppendToOutgoingContext(callCtx, "api_key", "dev-key")
		resp, err := tools.InvokeTool(callCtx, &proto.ExecuteToolRequest{
			SessionId:          sessionID,
			ToolName:           "add",
			Input:              `{"a":2,"b":3}`,
			WaitTimeoutSeconds: 20,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "invoke failed: %v\n", err)
			return cli.ExitError
		}
		fmt.Fprintf(out, "==> invoke add(a=2, b=3) -> %s (status %s)\n", resp.GetResult(), stringsTrimPrefix(resp.GetStatus().String(), "REQUEST_STATUS_"))
		fmt.Fprintln(out, "\ndemo complete. Next: toolplane init, then serve + provider serve + invoke in three terminals.")
		return cli.ExitOK
	}

	// --drill: show durable execution. Create a request, kill the
	// provider with it pending, then bring a fresh provider online — the
	// control plane holds the request, and the new provider claims and
	// finishes it.
	callCtx, callCancel := context.WithTimeout(ctx, 30*time.Second)
	defer callCancel()
	callCtx = metadata.AppendToOutgoingContext(callCtx, "api_key", "dev-key")
	resp, err := tools.InvokeTool(callCtx, &proto.ExecuteToolRequest{
		SessionId: sessionID,
		ToolName:  "add",
		Input:     `{"a":2,"b":3}`,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "invoke failed: %v\n", err)
		return cli.ExitError
	}
	pendingRequestID := resp.GetRequestId()
	fmt.Fprintf(out, "==> request %s pending; killing the provider...", pendingRequestID)
	_ = providerCmd.Process.Kill()
	_ = providerCmd.Wait()

	// The dead provider's machine still owns the session's tools. The
	// replacement provider cannot register a taken name, so release the
	// dead machine's registration first (what an operator's cleanup does).
	machines, mctx, cancelMachines, gconnMachines, err := dialMachines(conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconnMachines.Close()
	defer cancelMachines()
	machineList, err := machines.ListMachines(mctx, &proto.ListMachinesRequest{SessionId: sessionID})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	for _, m := range machineList.GetMachines() {
		_, err := machines.UnregisterMachine(mctx, &proto.UnregisterMachineRequest{
			SessionId: sessionID,
			MachineId: m.GetId(),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "unregister dead machine:", err)
			return cli.ExitError
		}
	}

	// A second provider takes over the session to finish the work.
	providerCmd2 := exec.Command(providerBin, providerArgs...)
	providerCmd2.Stdout = os.Stdout
	providerCmd2.Stderr = os.Stderr
	if err := providerCmd2.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer func() {
		_ = providerCmd2.Process.Kill()
		_ = providerCmd2.Wait()
	}()

	requests, reqCtx, cancelReq, gconnReq, err := dialRequests(conn)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return cli.ExitError
	}
	defer gconnReq.Close()
	defer cancelReq()

	deadline := time.Now().Add(60 * time.Second)
	for {
		req, err := requests.GetRequest(reqCtx, &proto.GetRequestRequest{
			SessionId: sessionID,
			RequestId: pendingRequestID,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "poll request: %v\n", err)
			return cli.ExitError
		}
		status := stringsTrimPrefix(req.GetStatus().String(), "REQUEST_STATUS_")
		fmt.Fprintf(out, "  status: %s\n", status)
		if status == "DONE" {
			fmt.Fprintf(out, "  result: %s\n", req.GetResult())
			fmt.Fprintln(out, "\ndrill complete: the provider died mid-request and the request still finished — that is durable execution.")
			return cli.ExitOK
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "drill: request did not settle within 60s")
			return cli.ExitError
		}
		time.Sleep(500 * time.Millisecond)
	}
}
