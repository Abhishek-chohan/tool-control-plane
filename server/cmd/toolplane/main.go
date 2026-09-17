// Command toolplane is the unified entry point for Toolplane: `serve`
// runs the gRPC control plane with the same lifecycle and flags as the
// standalone toolplane-server binary, and the client verbs speak api.v1
// over gRPC with a stable exit-code contract.
package main

import (
	"errors"
	"fmt"
	"os"

	"toolplane/internal/cli"
	"toolplane/pkg/service"
)

// version is overridden at build time with -ldflags "-X main.version=<tag>";
// dev builds report "dev".
var version = "dev"

// exitError carries a command's resolved exit code through cobra's RunE
// to main, which owns os.Exit.
type exitError struct{ code int }

func (e exitError) Error() string { return fmt.Sprintf("exit code %d", e.code) }

func main() {
	// HealthCheck reports the same identity `toolplane --version` prints,
	// so a client and the server it talks to can be compared directly.
	service.BuildVersion = version
	root := newRootCommand(version)
	if err := root.Execute(); err != nil {
		var coded exitError
		if errors.As(err, &coded) {
			os.Exit(coded.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCodeFor(err))
	}
}
