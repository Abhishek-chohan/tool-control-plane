package main

import (
	"github.com/spf13/cobra"
)

// newRootCommand assembles the toolplane command tree. Version is
// injected so `toolplane --version` reports the build, not the source.
func newRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "toolplane",
		Short: "Toolplane control plane and client verbs",
		Long: `Toolplane is a gRPC control plane for tool execution that outlives
the caller that started it: durable requests, provider machine ownership
with leases, bounded stream replay, and drain-safe rollouts.

  toolplane serve            run the control plane
  toolplane gateway serve    run the HTTP/JSON edge
  toolplane mcp serve        run the MCP edge

Client verbs (invoke, session, key, machine, request, task, tool) speak
api.v1 against a running server. Every command exits with a code mapping
the failure's gRPC status (2 invalid input, 3 not found, 4 denied,
5 precondition, 6 exhausted, 7 unavailable, 8 deadline, 9 cancelled) so
scripts branch without parsing output.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	root.AddCommand(newServeCommand())
	return root
}
