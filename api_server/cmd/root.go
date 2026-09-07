package main

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func new_root(listen func(addr string) error) *cobra.Command {
	port := "8080"
	run_listen := func(*cobra.Command, []string) error {
		return listen(listen_addr(port, os.Getenv("PORT")))
	}
	root := &cobra.Command{
		Use:           "api_server",
		Short:         "Retail graph HTTP surface",
		Long:          "Serve /healthz, /fold, /search, and /v1/graph. No args starts the server so the k8s Deployment stays unchanged.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE:          run_listen,
	}
	root.PersistentFlags().StringVar(&port, "port", "8080", "listen port (PORT env wins if set)")
	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Listen for HTTP (default)",
		Args:  cobra.NoArgs,
		RunE:  run_listen,
	})
	root.CompletionOptions.DisableDefaultCmd = true
	return root
}

func listen_addr(port_flag, port_env string) string {
	port := port_flag
	if port_env != "" {
		port = port_env
	}
	if port == "" {
		port = "8080"
	}
	return ":" + strings.TrimPrefix(port, ":")
}
