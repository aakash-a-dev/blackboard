package cmd

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/aakash-a-dev/blackboard/blackboard-cli/internal/config"
	"github.com/aakash-a-dev/blackboard/blackboard-cli/internal/logger"
	"github.com/aakash-a-dev/blackboard/blackboard-cli/internal/server"
	"github.com/aakash-a-dev/blackboard/blackboard-cli/internal/watcher"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve <file>",
	Short: "Parse a YAML file and start the mock HTTP server",
	Args:  cobra.ExactArgs(1),
	RunE:  runServe,
}

var (
	servePort      int
	serveHost      string
	serveWatch     bool
	serveLatency   int
	serveCORS      bool
	serveLogFormat string
)

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 4000, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "127.0.0.1", "Host/interface to bind")
	serveCmd.Flags().BoolVar(&serveWatch, "watch", true, "Hot-reload on YAML file save")
	serveCmd.Flags().IntVar(&serveLatency, "latency", 0, "Global artificial latency in ms")
	serveCmd.Flags().BoolVar(&serveCORS, "cors", true, "Emit permissive CORS headers")
	serveCmd.Flags().StringVar(&serveLogFormat, "log", "pretty", "Log format: pretty, json, silent")
}

func runServe(cmd *cobra.Command, args []string) error {
	path := args[0]

	logger.SetFormat(logger.Format(serveLogFormat))

	cfg, warnings, err := config.Load(path)
	if err != nil {
		return err
	}
	printWarnings(warnings)

	addr := fmt.Sprintf("%s:%d", serveHost, servePort)

	// Check port availability before printing the banner — decision #14
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is already in use. Use --port to specify a different port", servePort)
	}

	srv := server.New(serveCORS, serveLatency)

	printBanner(path)
	srv.Load(cfg)

	if serveWatch {
		go func() {
			watcher.Watch(path, func() { //nolint:errcheck
				newCfg, warns, loadErr := config.Load(path)
				if loadErr != nil {
					logger.Error("reload failed: " + loadErr.Error())
					return
				}
				printWarnings(warns)
				srv.Load(newCfg)
				logger.Info("  reloaded OK")
			})
		}()
	}

	logger.Info("")
	logger.Info("  " + strings.Repeat("─", 44))
	logger.Info(fmt.Sprintf("  Listening on http://%s", addr))
	logger.Info("  Press Ctrl+C to stop.")
	logger.Info("  " + strings.Repeat("─", 44))

	return http.Serve(ln, srv)
}

func printBanner(path string) {
	if logger.GetFormat() == logger.Silent {
		return
	}
	fmt.Printf("\n  \033[1mmockapi\033[0m  ·  %s\n\n", path)
}

func printWarnings(warnings []string) {
	for _, w := range warnings {
		logger.Warn(w)
	}
}
