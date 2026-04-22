package diagnostic

import (
	"fmt"
	"strings"

	"github.com/netscope/pkg/types"
)

// PrintDiagnosticReport prints a formatted diagnostic report
func PrintDiagnosticReport(report *types.DiagnosisReport, color bool) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                       网络诊断报告                                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Summary
	fmt.Printf("📊 诊断结果: %s\n\n", report.Summary)

	if len(report.Findings) == 0 {
		fmt.Println("✅ 未发现任何网络问题")
		return
	}

	// Findings by severity
	for _, finding := range report.Findings {
		printFinding(finding, color)
		fmt.Println()
	}
}

func printFinding(finding types.DiagnosisResult, color bool) {
	// Severity icon and color
	icon, colorCode := getSeverityInfo(finding.Severity)

	fmt.Printf("%s [%s] %s\n", icon, strings.ToUpper(finding.Code), finding.Title)
	fmt.Println(strings.Repeat("─", 70))

	// Message (may contain newlines)
	lines := strings.Split(finding.Message, "\n")
	for _, line := range lines {
		fmt.Printf("   %s\n", line)
	}

	fmt.Println()

	// Suggestion
	suggestionLines := strings.Split(finding.Suggestion, "\n")
	for _, line := range suggestionLines {
		fmt.Printf("   💡 建议: %s\n", line)
	}

	_ = colorCode // could use for terminal color codes
}

func getSeverityInfo(severity string) (string, string) {
	switch severity {
	case "critical":
		return "🔴", "31" // red
	case "warning":
		return "🟡", "33" // yellow
	case "info":
		return "🔵", "34" // blue
	default:
		return "⚪", "0"
	}
}
