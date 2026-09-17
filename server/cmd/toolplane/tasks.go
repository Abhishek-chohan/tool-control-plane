package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// dialTasks opens the control-plane connection for TasksService.
func dialTasks(conn cli.Connection) (proto.TasksServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewTasksServiceClient(gconn), ctx, cancel, gconn, nil
}

func newTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "List and inspect durable tasks",
	}
	cmd.AddCommand(newTaskListCommand())
	return cmd
}

func newTaskListCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list --session <id>",
		Short: "List a session's tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			tasks, sctx, cancel, gconn, err := dialTasks(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := tasks.ListTasks(sctx, &proto.ListTasksRequest{SessionId: session})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}

			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"ID", "TOOL", "STATUS", "CURRENT REQUEST"}
			rows := make([][]string, 0, len(resp.GetTasks()))
			list := make([]map[string]interface{}, 0, len(resp.GetTasks()))
			for _, t := range resp.GetTasks() {
				status := stringsTrimPrefix(t.GetStatus().String(), "TASK_STATUS_")
				rows = append(rows, []string{t.GetId(), t.GetToolName(), status, t.GetCurrentRequestId()})
				list = append(list, map[string]interface{}{
					"id": t.GetId(), "tool": t.GetToolName(), "status": status,
					"current_request_id": t.GetCurrentRequestId(),
				})
			}
			return r.Emit(header, rows, list)
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session to list tasks for")
	globals.Bind(cmd)
	return cmd
}
