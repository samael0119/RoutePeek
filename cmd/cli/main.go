package main

import (
	"fmt"
	"os"

	"github.com/routepeek/internal/discovery"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:   "routepeek",
	Short: "RoutePeek - Network topology visualization and diagnostic tool",
	Long: `RoutePeek helps you understand your machine's network configuration.
It shows all network interfaces, routes, DNS, VPN, and proxy settings
in an easy-to-understand format. Great for diagnosing network issues
like VPN affecting VM connections.`,
	Version: "0.1.0",
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan and display current network configuration",
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := discovery.GetNetworkSnapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			discovery.PrintJSON(snapshot)
		} else {
			discovery.PrintNetworkOverview(snapshot, !noColor)
		}
	},
}

var diagCmd = &cobra.Command{
	Use:   "diag",
	Short: "Run network diagnostics",
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := discovery.GetNetworkSnapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		report := discovery.RunDiagnostics(snapshot)

		if jsonOutput {
			discovery.PrintJSON(report)
		} else {
			discovery.PrintDiagnosticReport(report, !noColor)
		}
	},
}

var routeCmd = &cobra.Command{
	Use:   "routes",
	Short: "Show routing table",
	Run: func(cmd *cobra.Command, args []string) {
		routes, err := discovery.GetRoutes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			discovery.PrintJSON(routes)
		} else {
			discovery.PrintRoutes(routes, !noColor)
		}
	},
}

var interfacesCmd = &cobra.Command{
	Use:   "interfaces",
	Short: "Show all network interfaces",
	Aliases: []string{"iface", "if"},
	Run: func(cmd *cobra.Command, args []string) {
		interfaces, err := discovery.GetInterfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			discovery.PrintJSON(interfaces)
		} else {
			discovery.PrintInterfaces(interfaces, !noColor)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd, diagCmd, routeCmd, interfacesCmd)

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
