package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// dialRequests opens the control-plane connection for RequestsService.
func dialRequests(conn cli.Connection) (proto.RequestsServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewRequestsServiceClient(gconn), ctx, cancel, gconn, nil
}

func newRequestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "List, inspect, cancel, and follow requests",
	}
	cmd.AddCommand(
		newRequestListCommand(),
		newRequestCancelCommand(),
		newRequestLogsCommand(),
		newRequestStreamCommand(),
	)
	return cmd
}

func newRequestListCommand() *cobra.Command {
	var (
		session  string
		status   string
		toolName string
		pageSize int32
		pageTok  string
		globals  cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list --session <id>",
		Short: "List requests in a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			requests, sctx, cancel, gconn, err := dialRequests(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := requests.ListRequests(sctx, &proto.ListRequestsRequest{
				SessionId: session,
				Status:    statusFromName(status),
				ToolName:  toolName,
				PageSize:  pageSize,
				PageToken: pageTok,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}

			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"REQUEST ID", "TOOL", "STATUS"}
			rows := make([][]string, 0, len(resp.GetRequests()))
			list := make([]map[string]interface{}, 0, len(resp.GetRequests()))
			for _, req := range resp.GetRequests() {
				toolRow := req.GetToolName()
				statusRow := stringsTrimPrefix(req.GetStatus().String(), "REQUEST_STATUS_")
				rows = append(rows, []string{req.GetId(), toolRow, statusRow})
				list = append(list, map[string]interface{}{
					"id": req.GetId(), "tool": toolRow, "status": statusRow,
				})
			}
			if err := r.Emit(header, rows, list); err != nil {
				return exitError{code: cli.ExitError, msg: err.Error()}
			}
			if resp.GetPage() != nil && resp.GetPage().GetNextPageToken() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "next page: --page-token %s\n", resp.GetPage().GetNextPageToken())
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session to list requests for")
	cmd.Flags().StringVar(&status, "status", "", "filter by status (pending, claimed, running, done, failed, cancelled)")
	cmd.Flags().StringVar(&toolName, "tool", "", "filter by tool name")
	cmd.Flags().Int32Var(&pageSize, "page-size", 10, "requests per page")
	cmd.Flags().StringVar(&pageTok, "page-token", "", "continue from a previous page")
	globals.Bind(cmd)
	return cmd
}

func newRequestCancelCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "cancel <request-id>",
		Short: "Cancel a request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			requests, sctx, cancel, gconn, err := dialRequests(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := requests.CancelRequest(sctx, &proto.CancelRequestRequest{
				SessionId: session,
				RequestId: args[0],
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			if resp.GetSuccess() {
				fmt.Fprintf(cmd.OutOrStdout(), "request %s cancelled\n", args[0])
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session owning the request (required)")
	globals.Bind(cmd)
	return cmd
}

// newRequestLogsCommand replays the retained chunk window plus the final
// result — the tool's output after the fact.
func newRequestLogsCommand() *cobra.Command {
	var (
		session string
		fromSeq int32
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "logs <request-id>",
		Short: "Print a request's retained output chunks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			requests, sctx, cancel, gconn, err := dialRequests(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := requests.GetRequestChunks(sctx, &proto.GetRequestChunksRequest{
				SessionId: session,
				RequestId: args[0],
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			for i, chunk := range resp.GetChunks() {
				if int32(i) < fromSeq-1 {
					continue
				}
				fmt.Fprintln(cmd.OutOrStdout(), chunk)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session owning the request (required)")
	cmd.Flags().Int32Var(&fromSeq, "from-seq", 1, "first chunk sequence to print (1-based)")
	globals.Bind(cmd)
	return cmd
}

// newRequestStreamCommand follows a request live: new chunks print as
// they land (polled from the retained window), and the final status
// prints when the request settles.
func newRequestStreamCommand() *cobra.Command {
	var (
		session  string
		interval time.Duration
		globals  cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "stream -f <request-id>",
		Short: "Follow a request live: chunks and status transitions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			requests, sctx, cancel, gconn, err := dialRequests(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			printed := int32(0)
			lastStatus := ""
			deadline := time.Now().Add(10 * time.Minute)
			for time.Now().Before(deadline) {
				window, err := requests.GetRequestChunks(sctx, &proto.GetRequestChunksRequest{
					SessionId: session,
					RequestId: args[0],
				})
				if err != nil {
					fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
					return exitError{code: cli.ExitCodeFor(err)}
				}
				chunks := window.GetChunks()
				for ; int(printed) < len(chunks); printed++ {
					fmt.Fprintln(cmd.OutOrStdout(), chunks[printed])
				}

				req, err := requests.GetRequest(sctx, &proto.GetRequestRequest{
					SessionId: session,
					RequestId: args[0],
				})
				if err == nil {
					status := stringsTrimPrefix(req.GetStatus().String(), "REQUEST_STATUS_")
					if status != lastStatus {
						fmt.Fprintf(cmd.OutOrStdout(), "[status] %s\n", status)
						lastStatus = status
					}
					if isTerminalStatusName(status) {
						return nil
					}
				}
				time.Sleep(interval)
			}
			return exitError{code: cli.ExitDeadlineExceeded, msg: fmt.Sprintf(
				"request %s did not settle within the follow window", args[0])}
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session owning the request (required)")
	cmd.Flags().DurationVar(&interval, "poll-interval", 250*time.Millisecond, "poll cadence for new chunks and status")
	globals.Bind(cmd)
	return cmd
}

// isTerminalStatusName reports whether a status name settles a request.
func isTerminalStatusName(status string) bool {
	switch status {
	case "done", "failed", "cancelled":
		return true
	}
	return false
}
