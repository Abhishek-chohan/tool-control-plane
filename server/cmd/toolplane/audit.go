package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// newAuditCommand reads the durable audit trail (admin capability).
func newAuditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Read the durable audit trail",
	}
	cmd.AddCommand(newAuditListCommand())
	return cmd
}

func newAuditListCommand() *cobra.Command {
	var (
		session    string
		actorKeyID string
		event      string
		pageSize   int32
		pageToken  string
		globals    cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit events, newest first",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialSessions(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.ListAuditEvents(sctx, &proto.ListAuditEventsRequest{
				SessionId:  session,
				ActorKeyId: actorKeyID,
				Event:      event,
				PageSize:   pageSize,
				PageToken:  pageToken,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}

			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"ID", "EVENT", "SESSION", "ACTOR KEY"}
			rows := make([][]string, 0, len(resp.GetEvents()))
			list := make([]map[string]interface{}, 0, len(resp.GetEvents()))
			for _, e := range resp.GetEvents() {
				actor := e.GetActorKeyId()
				if actor == "" {
					actor = "(system)"
				}
				rows = append(rows, []string{
					fmt.Sprintf("%d", e.GetId()), e.GetEvent(), e.GetSessionId(), actor,
				})
				list = append(list, map[string]interface{}{
					"id": e.GetId(), "event": e.GetEvent(), "session_id": e.GetSessionId(),
					"actor_key_id": e.GetActorKeyId(), "details": e.GetDetails(),
				})
			}
			if err := r.Emit(header, rows, list); err != nil {
				return exitError{code: cli.ExitError, msg: err.Error()}
			}
			if resp.GetPage().GetNextPageToken() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "next page: --page-token %s\n", resp.GetPage().GetNextPageToken())
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "filter by session id")
	cmd.Flags().StringVar(&actorKeyID, "actor", "", "filter by acting API key id")
	cmd.Flags().StringVar(&event, "event", "", "filter by event type")
	cmd.Flags().Int32Var(&pageSize, "page-size", 50, "events per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "continue from a previous page")
	globals.Bind(cmd)
	return cmd
}
