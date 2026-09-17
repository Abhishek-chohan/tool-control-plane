package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// newInvokeCommand runs a tool: fire-and-forget by default (prints the
// request ID), blocking long-poll with --wait, or streaming with
// --stream.
func newInvokeCommand() *cobra.Command {
	var (
		session        string
		input          string
		wait           time.Duration
		stream         bool
		idempotencyKey string
		address        string
		apiKey         string
		timeout        time.Duration
		format         string
	)

	cmd := &cobra.Command{
		Use:   "invoke <tool>",
		Short: "Invoke a tool and print the result",
		Long: `Invoke a tool in a session.

  toolplane invoke add --session demo --input '{"a":2,"b":3}' --wait 30s

With --wait 0 (the default) the call is fire-and-forget: the request ID
prints and execution continues server-side. A positive --wait blocks
server-side until the tool reaches a terminal state or the wait elapses
(capped at 1h by the server; larger values are rejected). --stream
attaches to the tool's chunk stream and prints chunks as they land.
Output is a table by default or JSON with --format json.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			formatParsed, err := cli.ParseFormat(format)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return exitError{code: cli.ExitInvalidArgument}
			}
			conn := cli.NewConnection(address, apiKey, timeout)

			renderer := cmd.OutOrStdout()
			code, invokeErr := runInvoke(cmd.Context(), conn, renderer, invokeOptions{
				Tool:           strings.TrimSpace(args[0]),
				Session:        strings.TrimSpace(session),
				Input:          input,
				Wait:           wait,
				Stream:         stream,
				IdempotencyKey: idempotencyKey,
				Format:         formatParsed,
			})
			if invokeErr != nil {
				fmt.Fprintln(os.Stderr, teachError(invokeErr, conn))
			}
			return exitError{code: code}
		},
	}

	cmd.Flags().StringVar(&session, "session", "", "session to invoke in (required)")
	cmd.Flags().StringVar(&input, "input", "{}", "tool input as a JSON object string")
	cmd.Flags().DurationVar(&wait, "wait", 0, "block until terminal state or this duration elapses (0 = fire-and-forget)")
	cmd.Flags().BoolVar(&stream, "stream", false, "print stream chunks as they land instead of the final result")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "caller-chosen dedup key: retries return the original request")
	cmd.Flags().StringVar(&address, "address", cli.DefaultAddress, "control plane address")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (defaults to $"+cli.APIKeyEnvVar+")")
	cmd.Flags().DurationVar(&timeout, "timeout", cli.DefaultTimeout, "per-RPC deadline")
	cmd.Flags().StringVar(&format, "format", string(cli.FormatTable), "output format: table or json")

	return cmd
}

// invokeOptions carries the parsed invoke invocation.
type invokeOptions struct {
	Tool           string
	Session        string
	Input          string
	Wait           time.Duration
	Stream         bool
	IdempotencyKey string
	Format         cli.Format
}

// runInvoke performs one invocation and returns the process exit code.
// Transport errors teach the next command; RPC failures map onto the
// exit-code contract.
func runInvoke(ctx context.Context, conn cli.Connection, out io.Writer, opts invokeOptions) (int, error) {
	if opts.Tool == "" {
		return cli.ExitInvalidArgument, fmt.Errorf("a tool name is required: toolplane invoke <tool>")
	}
	if opts.Session == "" {
		return cli.ExitInvalidArgument, fmt.Errorf("--session is required: toolplane invoke %s --session <id>", opts.Tool)
	}
	// The wire takes input as a JSON string; reject non-JSON here with a
	// message naming the problem instead of a server-side INVALID_ARGUMENT.
	var inputCheck interface{}
	if err := json.Unmarshal([]byte(opts.Input), &inputCheck); err != nil {
		return cli.ExitInvalidArgument, fmt.Errorf("--input must be a JSON object: %v (got %q)", err, opts.Input)
	}

	gconn, err := conn.Dial()
	if err != nil {
		return cli.ExitUnavailable, err
	}
	defer gconn.Close()
	tool := proto.NewToolServiceClient(gconn)

	callCtx, cancel := conn.Call(ctx)
	defer cancel()

	req := &proto.ExecuteToolRequest{
		SessionId:          opts.Session,
		ToolName:           opts.Tool,
		Input:              opts.Input,
		IdempotencyKey:     opts.IdempotencyKey,
		WaitTimeoutSeconds: int32(opts.Wait.Seconds()),
	}

	if opts.Stream {
		return streamInvoke(tool, callCtx, out, opts)
	}

	resp, err := tool.InvokeTool(callCtx, req)
	if err != nil {
		return cli.ExitCodeFor(err), err
	}
	return printInvokeResult(out, opts, resp)
}

// printInvokeResult renders one terminal (or in-flight) response.
func printInvokeResult(out io.Writer, opts invokeOptions, resp *proto.ExecuteToolResponse) (int, error) {
	payload := map[string]interface{}{
		"request_id": resp.GetRequestId(),
		"status":     strings.TrimPrefix(resp.GetStatus().String(), "REQUEST_STATUS_"),
		"result":     resp.GetResult(),
		"error":      resp.GetError(),
	}
	r := cli.Renderer{Out: out, Format: opts.Format}
	if opts.Format == cli.FormatJSON {
		if err := r.JSON(payload); err != nil {
			return cli.ExitError, err
		}
	} else {
		if payload["status"] == "PENDING" || payload["status"] == "" {
			fmt.Fprintf(out, "request %s accepted (status %s); poll with:\n  toolplane request status %s --session %s\n",
				resp.GetRequestId(), payload["status"], resp.GetRequestId(), opts.Session)
			return cli.ExitOK, nil
		}
		if resp.GetError() != "" {
			fmt.Fprintf(out, "request %s failed: %s\n", resp.GetRequestId(), resp.GetError())
			return cli.ExitError, nil
		}
		fmt.Fprintf(out, "%s\n", resp.GetResult())
	}
	return cli.ExitOK, nil
}

// streamInvoke attaches to the chunk stream and prints chunks as they
// land, then the final status line.
func streamInvoke(tool proto.ToolServiceClient, ctx context.Context, out io.Writer, opts invokeOptions) (int, error) {
	stream, err := tool.StreamExecuteTool(ctx, &proto.ExecuteToolRequest{
		SessionId: opts.Session,
		ToolName:  opts.Tool,
		Input:     opts.Input,
	})
	if err != nil {
		return cli.ExitCodeFor(err), err
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cli.ExitCodeFor(err), err
		}
		fmt.Fprintln(out, chunk.GetChunk())
		if chunk.GetIsFinal() {
			break
		}
	}
	return cli.ExitOK, nil
}

// teachError upgrades common failures into next-step guidance.
func teachError(err error, conn cli.Connection) string {
	if cli.IsUnavailable(err) {
		return fmt.Sprintf("%v\nnothing is answering on %s — start a server with:\n  toolplane serve", err, conn.Address)
	}
	return err.Error()
}
