package main

import (
	"fmt"
	"os"

	"github.com/samael0119/RoutePeek/internal/discovery"
	"github.com/samael0119/RoutePeek/internal/i18n"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:   "routepeek",
	Short: i18n.T("cli_short"),
	Long: `RoutePeek helps you understand your machine's network configuration.
It shows all network interfaces, routes, DNS, VPN, and proxy settings
in an easy-to-understand format. Great for diagnosing network issues
like VPN affecting VM connections.`,
	Version: "0.1.0",
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: i18n.T("cli_scan"),
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
	Short: i18n.T("cli_diag"),
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
	Short: i18n.T("cli_routes"),
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
	Short: i18n.T("cli_iface"),
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
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, i18n.T("cli_json"))
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, i18n.T("cli_nocolor"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
