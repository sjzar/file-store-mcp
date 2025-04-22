package filestore

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/sjzar/file-store-mcp/internal/filestore"
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "debug")
	rootCmd.PersistentFlags().IntVar(&SSEPort, "sse-port", 0, "sse port")
	rootCmd.PersistentPreRun = initLog
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Err(err).Msg("command execution failed")
	}
}

var SSEPort int

var rootCmd = &cobra.Command{
	Use:     "file-store-mcp",
	Short:   "File Store MCP Server",
	Long:    `File Store MCP Server`,
	Example: `file-store-mcp`,
	Args:    cobra.MinimumNArgs(0),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Run: Root,
}

func Root(cmd *cobra.Command, args []string) {

	fs := filestore.New()

	if SSEPort > 0 {
		server := fs.NewSSEServer()
		defer func() { _ = server.Shutdown(cmd.Context()) }()
		log.Info().Msgf("SSE server started on port %d", SSEPort)
		if err := server.Start(fmt.Sprintf(":%d", SSEPort)); err != nil {
			log.Err(err).Msg("failed to start SSE server")
		}
		return
	}

	if err := fs.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
		log.Err(err).Msg("failed to run file store")
		return
	}

}
