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

// dialSessions opens the control-plane connection and returns a
// SessionsService client with the caller's deadline and key attached.
func dialSessions(conn cli.Connection) (proto.SessionsServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewSessionsServiceClient(gconn), ctx, cancel, gconn, nil
}

func newSessionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Create, list, delete, and invalidate sessions",
	}
	cmd.AddCommand(
		newSessionCreateCommand(),
		newSessionListCommand(),
		newSessionDeleteCommand(),
		newSessionInvalidateCommand(),
	)
	return cmd
}

func newSessionCreateCommand() *cobra.Command {
	var (
		userID      string
		name        string
		description string
		sessionID   string
		namespace   string
		globals     cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := globals.FormatParsed()
			if err != nil {
				return exitError{code: cli.ExitInvalidArgument, msg: err.Error()}
			}
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialSessions(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.CreateSession(sctx, &proto.CreateSessionRequest{
				UserId:      userID,
				Name:        name,
				Description: description,
				SessionId:   sessionID,
				Namespace:   namespace,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			sess := resp.GetSession()
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: format}
			return r.Emit(
				[]string{"SESSION ID", "NAME", "CREATED BY"},
				[][]string{{sess.GetId(), sess.GetName(), sess.GetCreatedBy()}},
				map[string]interface{}{"id": sess.GetId(), "name": sess.GetName(), "created_by": sess.GetCreatedBy()},
			)
		},
	}
	cmd.Flags().StringVar(&userID, "user", "", "user id owning the session")
	cmd.Flags().StringVar(&name, "name", "", "session name")
	cmd.Flags().StringVar(&description, "description", "", "session description")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "requested session id (generated when omitted)")
	cmd.Flags().StringVar(&namespace, "namespace", "", "optional namespace")
	globals.Bind(cmd)
	return cmd
}

func newSessionListCommand() *cobra.Command {
	var (
		userID  string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions for a user",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, ferr := globals.FormatParsed()
			if ferr != nil {
				return exitError{code: cli.ExitInvalidArgument, msg: ferr.Error()}
			}
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialSessions(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.ListSessions(sctx, &proto.ListSessionsRequest{UserId: userID})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: format}
			header := []string{"ID", "NAME", "CREATED BY"}
			rows := make([][]string, 0, len(resp.GetSessions()))
			list := make([]map[string]interface{}, 0, len(resp.GetSessions()))
			for _, sess := range resp.GetSessions() {
				rows = append(rows, []string{sess.GetId(), sess.GetName(), sess.GetCreatedBy()})
				list = append(list, map[string]interface{}{
					"id": sess.GetId(), "name": sess.GetName(), "created_by": sess.GetCreatedBy(),
				})
			}
			return r.Emit(header, rows, list)
		},
	}
	cmd.Flags().StringVar(&userID, "user", "", "user id to list sessions for")
	globals.Bind(cmd)
	return cmd
}

func newSessionDeleteCommand() *cobra.Command {
	var globals cli.Globals
	cmd := &cobra.Command{
		Use:   "delete <session-id>",
		Short: "Delete a session and its API keys",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialSessions(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.DeleteSession(sctx, &proto.DeleteSessionRequest{SessionId: args[0]})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			if resp.GetSuccess() {
				fmt.Fprintf(cmd.OutOrStdout(), "session %s deleted\n", args[0])
			}
			return nil
		},
	}
	globals.Bind(cmd)
	return cmd
}

func newSessionInvalidateCommand() *cobra.Command {
	var (
		reason  string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "invalidate <session-id>",
		Short: "Kill switch: revoke every live API key of a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialSessions(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.InvalidateSession(sctx, &proto.InvalidateSessionRequest{
				SessionId: args[0],
				Reason:    reason,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "invalidated %d api key(s) in session %s\n",
				resp.GetRevokedApiKeys(), args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "why the kill switch was pulled (recorded in the audit trail)")
	globals.Bind(cmd)
	return cmd
}
