package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"toolplane/internal/cli"
	"toolplane/internal/server"
	"toolplane/pkg/storage"
	proto "toolplane/proto"
)

// newDoctorCommand validates configuration, connectivity, and version
// skew without starting a server — the onboarding check: it names the
// exact setting or command that is missing.
func newDoctorCommand() *cobra.Command {
	var (
		globals     cli.Globals
		databaseURL string
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check configuration, connectivity, and version skew",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			failures := 0
			step := func(name, detail string, ok bool) {
				mark := "ok  "
				if !ok {
					mark = "FAIL"
					failures++
				}
				fmt.Fprintf(out, "[%s] %-14s %s\n", mark, name, detail)
			}

			// Configuration: the environment contract plus production gates.
			configErr := server.ValidateConfig()
			step("config", describeConfigState(configErr), configErr == nil)

			// Connectivity: can we reach the control plane at all?
			conn := globals.Connection()
			connectCtx, connectCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer connectCancel()
			connectErr := cli.WaitReady(connectCtx, conn, 250*time.Millisecond)
			step("connectivity", describeConnectivity(connectErr, conn), connectErr == nil)

			// Version skew: compare this binary against the server's build.
			var serverVersion string
			if connectErr == nil {
				serverVersion = probeServerVersion(conn)
			}
			skewOK := connectErr != nil || serverVersion == "" || serverVersion == version
			step("version", describeVersionSkew(version, serverVersion, connectErr == nil), skewOK)

			// Database: reachable and migrated? Only checkable with a URL —
			// OpenFromEnv runs migrations on connect, so a successful open
			// is the schema-freshness check.
			if databaseURL != "" {
				dbOK := checkDatabaseReachable(databaseURL)
				step("database", describeDatabase(dbOK), dbOK)
			}

			if failures > 0 {
				return exitError{code: cli.ExitError, msg: fmt.Sprintf("doctor found %d problem(s)", failures)}
			}
			fmt.Fprintln(out, "\nall checks passed")
			return nil
		},
	}
	cmd.Flags().StringVar(&databaseURL, "database-url", "", "check database reachability and migrations directly")
	globals.Bind(cmd)
	return cmd
}

func describeConfigState(err error) string {
	if err == nil {
		return "environment contract and production gates satisfied"
	}
	return err.Error()
}

func describeConnectivity(err error, conn cli.Connection) string {
	if err == nil {
		return "reachable at " + conn.Address
	}
	return fmt.Sprintf("unreachable at %s — start one with:\n  toolplane serve", conn.Address)
}

func describeVersionSkew(client, server string, connected bool) string {
	if !connected {
		return "skipped (server unreachable)"
	}
	if server == "" {
		return "server did not report a version (older build)"
	}
	if client == server {
		return fmt.Sprintf("client and server both %s", client)
	}
	return fmt.Sprintf("client %s vs server %s — versions differ", client, server)
}

func probeServerVersion(conn cli.Connection) string {
	gconn, err := conn.Dial()
	if err != nil {
		return ""
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
		return ""
	}
	return resp.GetVersion()
}

func checkDatabaseReachable(databaseURL string) bool {
	prev, had := os.LookupEnv("TOOLPLANE_DATABASE_URL")
	defer func() {
		if had {
			_ = os.Setenv("TOOLPLANE_DATABASE_URL", prev)
		} else {
			_ = os.Unsetenv("TOOLPLANE_DATABASE_URL")
		}
	}()
	if err := os.Setenv("TOOLPLANE_DATABASE_URL", databaseURL); err != nil {
		return false
	}
	store, err := storage.OpenFromEnv(context.Background(), nil)
	if err != nil {
		return false
	}
	return store.Close() == nil
}

func describeDatabase(ok bool) string {
	if ok {
		return "database reachable and schema ready"
	}
	return "database unreachable"
}
