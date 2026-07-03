package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/github/github-mcp-server/pkg/inventory"
	"github.com/github/github-mcp-server/pkg/notes"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "notes-mcp-server",
	Short:   "Notes MCP Server",
	Long:    "An MCP server that provides tools for managing in-memory notes.",
	Version: version,
}

var stdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Start stdio server",
	Long:  "Start a server that communicates via standard input/output using JSON-RPC messages.",
	RunE: func(_ *cobra.Command, _ []string) error {
		readOnly := viper.GetBool("read-only")
		logFilePath := viper.GetString("log-file")

		var slogHandler slog.Handler
		if logFilePath != "" {
			f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				return fmt.Errorf("failed to open log file: %w", err)
			}
			slogHandler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})
		} else {
			slogHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})
		}
		logger := slog.New(slogHandler)
		logger.Info("starting notes-mcp-server", "version", version, "readOnly", readOnly)

		// Build the in-memory store and register tools.
		store := notes.NewMemoryStore()
		tools := notes.AllTools(store)

		// Filter to read-only tools when --read-only is set.
		if readOnly {
			var ro []inventory.ServerTool
			for _, t := range tools {
				if t.IsReadOnly() {
					ro = append(ro, t)
				}
			}
			tools = ro
		}

		// Create the MCP server.
		s := mcp.NewServer(&mcp.Implementation{
			Name:    "notes-mcp-server",
			Title:   "Notes MCP Server",
			Version: version,
		}, nil)

		for i := range tools {
			t := tools[i]
			toolCopy := t.Tool
			s.AddTool(&toolCopy, t.Handler(nil))
		}

		// Handle OS signals for graceful shutdown.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		errC := make(chan error, 1)
		go func() {
			_, _ = fmt.Fprintln(os.Stderr, "Notes MCP Server running on stdio")
			errC <- s.Run(ctx, &mcp.IOTransport{Reader: os.Stdin, Writer: os.Stdout})
		}()

		select {
		case <-ctx.Done():
			logger.Info("shutting down")
		case err := <-errC:
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}
		}
		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetGlobalNormalizationFunc(wordSepNormalize)

	rootCmd.PersistentFlags().Bool("read-only", false, "Restrict the server to read-only operations")
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")

	_ = viper.BindPFlag("read-only", rootCmd.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))

	rootCmd.AddCommand(stdioCmd)
}

func initConfig() {
	viper.SetEnvPrefix("notes")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func wordSepNormalize(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
