package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"toolplane/internal/cli"
)

// scaffoldFile is one file the init command writes.
type scaffoldFile struct {
	path    string
	content string
}

const toolsPyTemplate = `"""Toolplane tools — served with: toolplane-provider serve tools.py"""

from toolplane.provider_registry import tool


@tool(name="add", description="Add two numbers")
def add(a: int, b: int) -> int:
    return a + b


@tool(name="greet", description="Greet someone")
def greet(name: str) -> str:
    return f"Hello, {name}!"
`

const toolplaneYamlTemplate = `# Base configuration for ` + "`toolplane serve`" + `.
# Flags override environment variables, which override this file.
# Unknown keys are a boot error. Values expand ${VAR} references.
#
# env: development                  # development | test | production
# auth:
#   mode: fixed                     # fixed | postgres (disabled needs opt-in)
#   fixed_api_key: dev-key
# storage:
#   mode: memory                    # memory | postgres
#   database_url: ${DATABASE_URL}
#   # database_url_file: /run/secrets/db_url
# server:
#   port: 9001
#   metrics_listen: "127.0.0.1:0"
#   cert_file: /etc/toolplane/server.crt
#   key_file: /etc/toolplane/server.key
`

const envTemplate = `# Toolplane development settings — export these before running commands:
#   export $(grep -v '^#' .env | xargs)
TOOLPLANE_ENV_MODE=development
TOOLPLANE_AUTH_MODE=fixed
TOOLPLANE_AUTH_FIXED_API_KEY=dev-key
TOOLPLANE_STORAGE_MODE=memory
TOOLPLANE_API_KEY=dev-key
`

// newInitCommand scaffolds a working directory so the newcomer loop runs
// with zero flags: init → serve → invoke.
func newInitCommand() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a tools file, config, and env in this directory",
		Long: `Scaffold a working directory for Toolplane:

  tools.py        tools served by: toolplane-provider serve tools.py
  toolplane.yaml  base configuration for: toolplane serve --config ...
  .env            development settings (fixed dev key, in-memory storage)

Existing files are never overwritten. The full loop afterwards:

  toolplane serve                              # terminal 1
  toolplane-provider serve tools.py            # terminal 2
  toolplane invoke add --session <id> --input '{"a":2,"b":3}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				dir = "."
			}
			scaffolds := []scaffoldFile{
				{filepath.Join(dir, "tools.py"), toolsPyTemplate},
				{filepath.Join(dir, "toolplane.yaml"), toolplaneYamlTemplate},
				{filepath.Join(dir, ".env"), envTemplate},
			}

			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return exitError{code: cli.ExitError}
			}

			var existing []string
			for _, file := range scaffolds {
				if _, err := os.Stat(file.path); err == nil {
					existing = append(existing, file.path)
				}
			}
			if len(existing) > 0 {
				return exitError{code: cli.ExitInvalidArgument, msg: fmt.Sprintf(
					"refusing to overwrite existing file(s): %s — remove them or init in another directory",
					strings.Join(existing, ", "))}
			}

			for _, file := range scaffolds {
				if err := os.WriteFile(file.path, []byte(file.content), 0o600); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return exitError{code: cli.ExitError}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", file.path)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nNext steps:\n  1. toolplane serve\n"+
				"  2. toolplane-provider serve tools.py --session demo\n"+
				"  3. toolplane invoke add --session demo --input '{\"a\":2,\"b\":3}' --wait 30s\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "directory to scaffold into")
	return cmd
}
