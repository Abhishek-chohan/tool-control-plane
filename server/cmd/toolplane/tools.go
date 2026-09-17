package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// dialToolsCLI opens the control-plane connection for tool verbs.
func dialToolsCLI(conn cli.Connection) (proto.ToolServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewToolServiceClient(gconn), ctx, cancel, gconn, nil
}

func newToolsCommand() *cobra.Command {
	var (
		session string
		openai  bool
		globals cli.Globals
	)
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "List registered tools, optionally for agent function-calling",
	}
	list := &cobra.Command{
		Use:   "list --session <id>",
		Short: "List a session's registered tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := globals.Connection()
			tools, sctx, cancel, gconn, err := dialToolsCLI(conn)
			if err != nil {
				return exitError{code: cli.ExitUnavailable, msg: cli.TeachError(err, conn)}
			}
			defer gconn.Close()
			defer cancel()

			resp, err := tools.ListTools(sctx, &proto.ListToolsRequest{SessionId: session})
			if err != nil {
				fmt.Fprintln(os.Stderr, cli.TeachError(err, conn))
				return exitError{code: cli.ExitCodeFor(err)}
			}

			// OpenAI function-calling shape: the schema renders as a parsed
			// JSON object, never a double-encoded string.
			if openai {
				type functionDef struct {
					Name        string      `json:"name"`
					Description string      `json:"description"`
					Parameters  interface{} `json:"parameters"`
				}
				type openAITool struct {
					Type     string      `json:"type"`
					Function functionDef `json:"function"`
				}
				var payload []openAITool
				for _, t := range resp.GetTools() {
					var params interface{}
					if err := json.Unmarshal([]byte(t.GetSchema()), &params); err != nil {
						params = map[string]interface{}{"type": "object"}
					}
					payload = append(payload, openAITool{
						Type: "function",
						Function: functionDef{
							Name:        t.GetName(),
							Description: t.GetDescription(),
							Parameters:  params,
						},
					})
				}
				encoded, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return exitError{code: cli.ExitError, msg: err.Error()}
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
				return nil
			}

			r := cli.Renderer{Out: cmd.OutOrStdout(), Format: globals.FormatParsedOrDefault()}
			header := []string{"ID", "NAME", "MACHINE"}
			rows := make([][]string, 0, len(resp.GetTools()))
			list := make([]map[string]interface{}, 0, len(resp.GetTools()))
			for _, t := range resp.GetTools() {
				rows = append(rows, []string{t.GetId(), t.GetName(), t.GetMachineId()})
				list = append(list, map[string]interface{}{
					"id": t.GetId(), "name": t.GetName(), "machine_id": t.GetMachineId(),
				})
			}
			return r.Emit(header, rows, list)
		},
	}
	list.Flags().StringVar(&session, "session", "", "session to list tools for (required)")
	list.Flags().BoolVar(&openai, "openai", false, "render tools in OpenAI function-calling format")
	globals.Bind(list)
	cmd.AddCommand(list)
	return cmd
}
