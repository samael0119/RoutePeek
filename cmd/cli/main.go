package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/samael0119/RoutePeek/internal/connectivity"
	"github.com/samael0119/RoutePeek/internal/discovery"
	"github.com/samael0119/RoutePeek/internal/i18n"
	reportpkg "github.com/samael0119/RoutePeek/internal/report"
	"github.com/samael0119/RoutePeek/pkg/types"
	"github.com/spf13/cobra"
)

var (
	jsonOutput   bool
	noColor      bool
	reportFormat string
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
	Use:     "interfaces",
	Short:   i18n.T("cli_iface"),
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

var checkCmd = &cobra.Command{
	Use:   "check [target...]",
	Short: i18n.T("cli_check"),
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := discovery.GetNetworkSnapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		matrix := connectivity.Check(context.Background(), snapshot, args, "")
		if jsonOutput {
			discovery.PrintJSON(matrix)
			return
		}
		printConnectivityMatrix(matrix)
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: i18n.T("cli_report"),
	Run: func(cmd *cobra.Command, args []string) {
		format := strings.ToLower(strings.TrimSpace(reportFormat))
		if format == "" {
			format = "markdown"
		}
		if format != "markdown" && format != "json" {
			fmt.Fprintf(os.Stderr, "Error: --format must be markdown or json\n")
			os.Exit(1)
		}

		payload, err := reportpkg.Build(context.Background(), "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if format == "json" {
			discovery.PrintJSON(payload)
			return
		}
		fmt.Print(reportpkg.Markdown(payload, ""))
	},
}

func init() {
	rootCmd.Version = reportpkg.Version
	rootCmd.AddCommand(scanCmd, diagCmd, routeCmd, interfacesCmd, checkCmd, reportCmd)

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, i18n.T("cli_json"))
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, i18n.T("cli_nocolor"))
	reportCmd.Flags().StringVar(&reportFormat, "format", "markdown", i18n.T("cli_report_format"))
}

func printConnectivityMatrix(matrix *types.ConnectivityReport) {
	fmt.Printf("RoutePeek Connectivity Matrix\n")
	fmt.Printf("%s\n\n", matrix.Summary)
	fmt.Printf("%-18s %-8s %-8s %-8s %-10s %s\n", "TARGET", "DNS", "TCP", "HTTP", "PATH", "CONCLUSION")
	for _, row := range matrix.Targets {
		fmt.Printf("%-18s %-8s %-8s %-8s %-10s %s\n",
			truncate(row.Target.Name, 18),
			row.DNS.Status,
			row.TCP.Status,
			row.HTTP.Status,
			row.Path.Status,
			row.Conclusion,
		)
	}
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	if max <= 1 {
		return value[:max]
	}
	return value[:max-1] + "."
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
