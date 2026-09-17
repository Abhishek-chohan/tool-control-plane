package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// newStatusCommand prints one screen of truth about the control plane:
// reachability, build identity, and resolved storage mode.
func newStatusCommand() *cobra.Command {
	var globals cli.Globals
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show control-plane reachability, version, and storage mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			gconn, err := conn.Dial()
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			tools := proto.NewToolServiceClient(gconn)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if conn.APIKey != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "api_key", conn.APIKey)
			}
			resp, err := tools.HealthCheck(ctx, &proto.HealthCheckRequest{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "server at %s is unreachable: %v\nstart one with:\n  toolplane serve\n",
					conn.Address, err)
				return exitError{code: cli.ExitUnavailable}
			}

			storage := resp.GetStorage()
			if storage == "" {
				storage = "(server predates storage reporting)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "control plane: reachable at %s\n  version: %s\n  storage: %s\n",
				conn.Address, resp.GetVersion(), storage)
			return nil
		},
	}
	globals.Bind(cmd)
	return cmd
}
