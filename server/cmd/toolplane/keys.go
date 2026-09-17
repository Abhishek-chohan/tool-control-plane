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

// dialKeys opens the control-plane connection for API key verbs.
func dialKeys(conn cli.Connection) (proto.SessionsServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	return dialSessions(conn)
}

func newKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Create, list, revoke, and test API keys",
	}
	cmd.AddCommand(newKeyCreateCommand(), newKeyListCommand(), newKeyRevokeCommand(), newKeyTestCommand())
	return cmd
}

func newKeyCreateCommand() *cobra.Command {
	var (
		session      string
		name         string
		capabilities []string
		allowedTools []string
		globals      cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "create --session <id>",
		Short: "Create an API key (the secret prints exactly once)",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialKeys(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.CreateApiKey(sctx, &proto.CreateApiKeyRequest{
				SessionId:    session,
				Name:         name,
				Capabilities: capabilities,
				AllowedTools: allowedTools,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			// The secret is served exactly once; render it in both formats
			// (JSON consumers capture it now — it never comes back).
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			if err := r.Emit(
				[]string{"KEY ID", "SECRET (copy now — shown once)", "CAPABILITIES"},
				[][]string{{resp.GetId(), resp.GetKey(), joinCapabilities(resp.GetCapabilities())}},
				map[string]interface{}{"id": resp.GetId(), "key": resp.GetKey(), "capabilities": resp.GetCapabilities()},
			); err != nil {
				return exitError{code: cli.ExitError, msg: err.Error()}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session to mint the key for (required)")
	cmd.Flags().StringVar(&name, "name", "", "key name")
	cmd.Flags().StringSliceVar(&capabilities, "capabilities", nil, "capabilities: read, invoke, provide, admin")
	cmd.Flags().StringSliceVar(&allowedTools, "allowed-tools", nil, "restrict the key to these tool names")
	globals.Bind(cmd)
	return cmd
}

func newKeyListCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list --session <id>",
		Short: "List a session's API keys (secrets never shown)",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialKeys(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.ListApiKeys(sctx, &proto.ListApiKeysRequest{SessionId: session})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"ID", "NAME", "PREVIEW", "REVOKED"}
			rows := make([][]string, 0, len(resp.GetApiKeys()))
			list := make([]map[string]interface{}, 0, len(resp.GetApiKeys()))
			for _, key := range resp.GetApiKeys() {
				revoked := "no"
				if key.GetRevokedAt() != nil {
					revoked = "yes"
				}
				rows = append(rows, []string{key.GetId(), key.GetName(), key.GetKeyPreview(), revoked})
				list = append(list, map[string]interface{}{
					"id": key.GetId(), "name": key.GetName(), "preview": key.GetKeyPreview(), "revoked": revoked,
				})
			}
			return r.Emit(header, rows, list)
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session to list keys for")
	globals.Bind(cmd)
	return cmd
}

func newKeyRevokeCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "revoke --session <id> <key-id>",
		Short: "Revoke an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			sessions, sctx, cancel, gconn, err := dialKeys(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := sessions.RevokeApiKey(sctx, &proto.RevokeApiKeyRequest{
				SessionId: session,
				KeyId:     args[0],
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			if resp.GetSuccess() {
				fmt.Fprintf(cmd.OutOrStdout(), "api key %s revoked\n", args[0])
			}
			return nil
		},
	}
	globals.Bind(cmd)
	return cmd
}

// newKeyTestCommand proves a key works: it runs an authenticated health
// probe and reports the outcome — catching typos and revocations without
// touching business RPCs.
func newKeyTestCommand() *cobra.Command {
	var globals cli.Globals
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Verify a key authenticates (runs an authenticated health probe)",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			if conn.APIKey == "" {
				return exitError{code: cli.ExitInvalidArgument, msg: cli.ErrNoConfig.Error()}
			}
			tools, sctx, cancel, gconn, err := dialTools(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := tools.HealthCheck(sctx, &proto.HealthCheckRequest{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "key rejected: %v\n", cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "key accepted — server %s (version %s)\n", conn.Address, resp.GetVersion())
			return nil
		},
	}
	globals.Bind(cmd)
	return cmd
}

func joinCapabilities(caps []string) string {
	out := ""
	for i, c := range caps {
		if i > 0 {
			out += ","
		}
		out += c
	}
	return out
}
