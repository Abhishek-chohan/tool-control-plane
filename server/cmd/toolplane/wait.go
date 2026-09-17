package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
)

// newWaitCommand blocks on server readiness — for systemd ExecStartPre,
// compose healthchecks, and scripts that must not act before the control
// plane answers.
func newWaitCommand() *cobra.Command {
	var (
		waitFor string
		timeout time.Duration
		address string
		apiKey  string
	)

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Block until the control plane is ready",
		Long: `Block until the control plane answers a health probe.

  toolplane wait --for ready --timeout 30s

Exits 0 once ready; exits with the failure's code on timeout (with the
toolplane serve next step).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if waitFor != "ready" {
				return exitError{code: cli.ExitInvalidArgument, msg: fmt.Sprintf(
					"--for %q is not supported; the only condition is ready", waitFor)}
			}
			conn := cli.NewConnection(address, apiKey, timeout)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := cli.WaitReady(ctx, conn, cli.DefaultReadyPollInterval); err != nil {
				return exitError{code: cli.ExitCodeFor(err), msg: cli.TeachError(
					fmt.Errorf("server at %s not ready within %s: %v", conn.Address, timeout, err), conn)}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "server at %s is ready\n", conn.Address)
			return nil
		},
	}

	cmd.Flags().StringVar(&waitFor, "for", "ready", "condition to wait for (ready)")
	cmd.Flags().DurationVar(&timeout, "timeout", cli.DefaultTimeout, "give up after this duration")
	cmd.Flags().StringVar(&address, "address", cli.DefaultAddress, "control plane address")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (defaults to $"+cli.APIKeyEnvVar+")")

	return cmd
}
