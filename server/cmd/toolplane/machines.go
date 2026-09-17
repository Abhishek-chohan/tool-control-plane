package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

func newMachinesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machine",
		Short: "List, inspect, and drain provider machines",
	}
	cmd.AddCommand(newMachineListCommand(), newMachineStatusCommand(), newMachineDrainCommand())
	return cmd
}

func newMachineListCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "list --session <id>",
		Short: "List a session's machines with heartbeat age",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			machines, sctx, cancel, gconn, err := dialMachines(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := machines.ListMachines(sctx, &proto.ListMachinesRequest{SessionId: session})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"ID", "SDK", "LAST PING (age)", "DRAINING"}
			rows := make([][]string, 0, len(resp.GetMachines()))
			list := make([]map[string]interface{}, 0, len(resp.GetMachines()))
			for _, m := range resp.GetMachines() {
				age := "never"
				if m.GetLastPingAt() != nil {
					age = time.Since(m.GetLastPingAt().AsTime()).Round(time.Second).String()
				}
				draining := "no"
				if m.GetDraining() {
					draining = "yes"
				}
				rows = append(rows, []string{m.GetId(), m.GetSdkVersion() + "/" + m.GetSdkLanguage(), age, draining})
				list = append(list, map[string]interface{}{
					"id": m.GetId(), "sdk": m.GetSdkVersion(), "last_ping_age": age, "draining": draining,
				})
			}
			return r.Emit(header, rows, list)
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session to list machines for")
	globals.Bind(cmd)
	return cmd
}

func newMachineStatusCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "status <machine-id>",
		Short: "Inspect one machine: heartbeat, drain state, identity",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			machines, sctx, cancel, gconn, err := dialMachines(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			m, err := machines.GetMachine(sctx, &proto.GetMachineRequest{SessionId: session, MachineId: args[0]})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			age := "never"
			if m.GetLastPingAt() != nil {
				age = time.Since(m.GetLastPingAt().AsTime()).Round(time.Second).String()
			}
			draining := "no"
			if m.GetDraining() {
				draining = "yes"
			}
			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			return r.Emit(
				[]string{"ID", "SDK", "LAST PING (age)", "DRAINING"},
				[][]string{{m.GetId(), m.GetSdkVersion() + "/" + m.GetSdkLanguage(), age, draining}},
				map[string]interface{}{
					"id": m.GetId(), "sdk": m.GetSdkVersion(), "last_ping_age": age, "draining": draining,
				},
			)
		},
	}
	globals.Bind(cmd)
	return cmd
}

func newMachineDrainCommand() *cobra.Command {
	var (
		session string
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "drain <machine-id>",
		Short: "Drain a machine: stop new work, finish in-flight requests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			machines, sctx, cancel, gconn, err := dialMachines(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := machines.DrainMachine(sctx, &proto.DrainMachineRequest{SessionId: session, MachineId: args[0]})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "machine %s drained (success=%v)\n", args[0], resp.GetDrained())
			return nil
		},
	}
	globals.Bind(cmd)
	return cmd
}
