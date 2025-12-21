package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santiagolertora/blc-perf-analyzer/internal/heatmap"
	"github.com/santiagolertora/blc-perf-analyzer/internal/parser"
)

// AnalysisResult contains the analysis results
type AnalysisResult struct {
	TopFunctions []FunctionStats `json:"top_functions"`
	Summary      SummaryStats    `json:"summary"`
}

// FunctionStats contains statistics for a single function
type FunctionStats struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"` // "userland", "kernel", "unknown"
	Percentage      float64 `json:"percentage"`
	TotalSamples    int     `json:"total_samples"`
	SelfSamples     int     `json:"self_samples"`
	ChildrenSamples int     `json:"children_samples"`
}

// SummaryStats contains summary statistics
type SummaryStats struct {
	TotalSamples    int     `json:"total_samples"`
	UserlandPercent float64 `json:"userland_percent"`
	KernelPercent   float64 `json:"kernel_percent"`
	UnknownPercent  float64 `json:"unknown_percent"`
	CaptureDuration int     `json:"capture_duration"`
	ProcessName     string  `json:"process_name"`
	PID             int     `json:"pid"`
}

// GenerateReport generates a complete analysis report including flamegraph
func GenerateReport(perfDataPath, outputDir string, processName string, pid int, duration int, generateHeatmapFlag bool, heatmapWindowSize float64) error {
	// 1. Generate flamegraph
	if err := generateFlamegraph(perfDataPath, outputDir); err != nil {
		return fmt.Errorf("error generating flamegraph: %v", err)
	}

	// 2. Generate perf report
	if err := generatePerfReport(perfDataPath, outputDir); err != nil {
		return fmt.Errorf("error generating perf report: %v", err)
	}

	// 3. Parse perf script output for advanced analysis
	samples, err := parsePerfScriptData(perfDataPath)
	if err != nil {
		fmt.Printf("Warning: Could not parse perf script for advanced analysis: %v\n", err)
		samples = []*parser.Sample{} // Continue with empty samples
	}

	// 4. Generate heatmap if requested and samples available
	if generateHeatmapFlag && len(samples) > 0 {
		fmt.Println("Generating interactive heatmap...")
		if err := heatmap.GenerateHeatmap(samples, outputDir, processName, pid, heatmapWindowSize); err != nil {
			fmt.Printf("Warning: Could not generate heatmap: %v\n", err)
		}
	}

	// 5. Generate summary with parsed data
	if err := generateSummary(perfDataPath, outputDir, processName, pid, duration, samples); err != nil {
		return fmt.Errorf("error generating summary: %v", err)
	}

	return nil
}

func generateFlamegraph(perfDataPath, outputDir string) error {
	fmt.Println("Generating flamegraph...")

	// First, generate the folded stack
	foldedPath := filepath.Join(outputDir, "perf.folded")
	fmt.Println("Running perf script to generate stack traces...")
	cmd := exec.Command("perf", "script", "-i", perfDataPath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running perf script: %v", err)
	}

	// Process the output to create folded stacks
	fmt.Println("Processing stack traces...")
	foldedStacks := processPerfOutput(string(output))
	if err := os.WriteFile(foldedPath, []byte(foldedStacks), 0644); err != nil {
		return fmt.Errorf("error writing folded stacks: %v", err)
	}

	// Check if flamegraph.pl is available
	fmt.Println("Checking for flamegraph.pl...")
	flamegraphPath, err := exec.LookPath("flamegraph.pl")
	if err != nil {
		fmt.Println("flamegraph.pl not found, downloading...")
		// Try to download flamegraph.pl
		if err := downloadFlamegraph(outputDir); err != nil {
			return fmt.Errorf("error downloading flamegraph.pl: %v", err)
		}
		flamegraphPath = filepath.Join(outputDir, "flamegraph.pl")
	}

	// Generate the flamegraph
	fmt.Println("Generating flamegraph visualization...")
	cmd = exec.Command(flamegraphPath, "--title", "CPU Flame Graph", "--countname", "samples", foldedPath)
	output, err = cmd.Output()
	if err != nil {
		// If the command fails, try to get more detailed error information
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("error generating flamegraph: %v\nstderr: %s", err, exitErr.Stderr)
		}
		return fmt.Errorf("error generating flamegraph: %v", err)
	}

	// Save the flamegraph
	flamegraphPath = filepath.Join(outputDir, "flamegraph.svg")
	fmt.Println("Saving flamegraph to", flamegraphPath)
	if err := os.WriteFile(flamegraphPath, output, 0644); err != nil {
		return fmt.Errorf("error saving flamegraph: %v", err)
	}

	fmt.Println("Flamegraph generation complete!")
	return nil
}

func generatePerfReport(perfDataPath, outputDir string) error {
	// Generate perf report
	cmd := exec.Command("perf", "report", "-i", perfDataPath, "--stdio")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error generating perf report: %v", err)
	}

	// Save the report
	reportPath := filepath.Join(outputDir, "perf-report.txt")
	if err := os.WriteFile(reportPath, output, 0644); err != nil {
		return fmt.Errorf("error saving perf report: %v", err)
	}

	return nil
}

func generateSummary(perfDataPath, outputDir, processName string, pid int, duration int, samples []*parser.Sample) error {
	// Generate perf report for analysis
	cmd := exec.Command("perf", "report", "-i", perfDataPath, "--stdio")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error generating perf report for analysis: %v", err)
	}

	// Parse the report using both old and new methods
	stats := parsePerfReport(string(output), samples)

	// Create summary
	summary := SummaryStats{
		TotalSamples:    stats.Summary.TotalSamples,
		UserlandPercent: stats.Summary.UserlandPercent,
		KernelPercent:   stats.Summary.KernelPercent,
		UnknownPercent:  stats.Summary.UnknownPercent,
		CaptureDuration: duration,
		ProcessName:     processName,
		PID:             pid,
	}

	// Save summary as JSON
	summaryJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling summary: %v", err)
	}

	summaryPath := filepath.Join(outputDir, "summary.json")
	if err := os.WriteFile(summaryPath, summaryJSON, 0644); err != nil {
		return fmt.Errorf("error saving summary: %v", err)
	}

	// Save human-readable summary
	summaryText := generateSummaryText(summary, stats.TopFunctions)
	summaryTextPath := filepath.Join(outputDir, "summary.txt")
	if err := os.WriteFile(summaryTextPath, []byte(summaryText), 0644); err != nil {
		return fmt.Errorf("error saving summary text: %v", err)
	}

	return nil
}

func downloadFlamegraph(outputDir string) error {
	// Download flamegraph.pl from GitHub
	cmd := exec.Command("curl", "-L", "https://raw.githubusercontent.com/brendangregg/FlameGraph/master/flamegraph.pl", "-o", filepath.Join(outputDir, "flamegraph.pl"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error downloading flamegraph.pl: %v", err)
	}

	// Make it executable
	cmd = exec.Command("chmod", "+x", filepath.Join(outputDir, "flamegraph.pl"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error making flamegraph.pl executable: %v", err)
	}

	return nil
}

func processPerfOutput(output string) string {
	// Process perf script output to create folded stacks
	var folded strings.Builder
	lines := strings.Split(output, "\n")

	// Track unique stacks to avoid duplicates
	stackCounts := make(map[string]int)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the perf script output
		// Example line: "process 1234 [000] 123.456: cpu-clock: ..."
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// Extract the stack trace (everything after the timestamp)
		stackStart := -1
		for i, part := range parts {
			if strings.HasSuffix(part, ":") {
				stackStart = i + 1
				break
			}
		}
		if stackStart == -1 || stackStart >= len(parts) {
			continue
		}

		// Create the folded stack
		stack := strings.Join(parts[stackStart:], ";")
		stackCounts[stack]++
	}

	// Write the folded stacks
	for stack, count := range stackCounts {
		folded.WriteString(fmt.Sprintf("%s %d\n", stack, count))
	}

	return folded.String()
}

func parsePerfReport(report string, samples []*parser.Sample) *AnalysisResult {
	result := &AnalysisResult{
		TopFunctions: make([]FunctionStats, 0),
		Summary: SummaryStats{
			TotalSamples:    len(samples),
			UserlandPercent: 0,
			KernelPercent:   0,
			UnknownPercent:  0,
		},
	}

	if len(samples) == 0 {
		return result
	}

	// Count by function and category
	functionCounts := make(map[string]*FunctionStats)
	var kernelCount, userlandCount, unknownCount int

	for _, sample := range samples {
		if topFrame := sample.GetTopFrame(); topFrame != nil {
			key := topFrame.Symbol

			if _, exists := functionCounts[key]; !exists {
				funcType := "unknown"
				if topFrame.IsKernel {
					funcType = "kernel"
				} else if topFrame.IsUserland {
					funcType = "userland"
				}

				functionCounts[key] = &FunctionStats{
					Name:         topFrame.Symbol,
					Type:         funcType,
					TotalSamples: 0,
					SelfSamples:  0,
				}
			}

			functionCounts[key].SelfSamples++
			functionCounts[key].TotalSamples++

			// Count categories
			if topFrame.IsKernel {
				kernelCount++
			} else if topFrame.IsUserland {
				userlandCount++
			} else {
				unknownCount++
			}
		}
	}

	// Calculate percentages
	totalSamples := float64(len(samples))
	if totalSamples > 0 {
		result.Summary.KernelPercent = float64(kernelCount) / totalSamples * 100
		result.Summary.UserlandPercent = float64(userlandCount) / totalSamples * 100
		result.Summary.UnknownPercent = float64(unknownCount) / totalSamples * 100
	}

	// Convert to slice and calculate percentages
	for _, stats := range functionCounts {
		stats.Percentage = float64(stats.SelfSamples) / totalSamples * 100
		result.TopFunctions = append(result.TopFunctions, *stats)
	}

	// Sort by total samples descending
	sort.Slice(result.TopFunctions, func(i, j int) bool {
		return result.TopFunctions[i].TotalSamples > result.TopFunctions[j].TotalSamples
	})

	return result
}

// parsePerfScriptData executes perf script and parses the output
func parsePerfScriptData(perfDataPath string) ([]*parser.Sample, error) {
	fmt.Println("Parsing perf script output for detailed analysis...")
	
	cmd := exec.Command("perf", "script", "-i", perfDataPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running perf script: %v", err)
	}

	samples, err := parser.ParsePerfScript(string(output))
	if err != nil {
		return nil, fmt.Errorf("error parsing perf script: %v", err)
	}

	fmt.Printf("Parsed %d samples from perf data\n", len(samples))
	return samples, nil
}

func generateSummaryText(summary SummaryStats, topFunctions []FunctionStats) string {
	var text strings.Builder

	text.WriteString("Performance Analysis Summary\n")
	text.WriteString("==========================\n\n")

	text.WriteString(fmt.Sprintf("Process: %s (PID: %d)\n", summary.ProcessName, summary.PID))
	text.WriteString(fmt.Sprintf("Duration: %d seconds\n", summary.CaptureDuration))
	text.WriteString(fmt.Sprintf("Total Samples: %d\n\n", summary.TotalSamples))

	text.WriteString("Time Distribution:\n")
	text.WriteString(fmt.Sprintf("- Userland: %.2f%%\n", summary.UserlandPercent))
	text.WriteString(fmt.Sprintf("- Kernel: %.2f%%\n", summary.KernelPercent))
	text.WriteString(fmt.Sprintf("- Unknown: %.2f%%\n\n", summary.UnknownPercent))

	text.WriteString("Top Functions:\n")
	unknownCount := 0
	for i, fn := range topFunctions {
		if i >= 10 { // Show only top 10
			break
		}
		text.WriteString(fmt.Sprintf("%d. %s (%.2f%%)\n", i+1, fn.Name, fn.Percentage))
		if fn.Name == "[unknown]" || strings.Contains(fn.Name, "unknown") {
			unknownCount++
		}
	}

	// Add recommendations if many unknowns
	if len(topFunctions) > 0 && topFunctions[0].Name == "[unknown]" && topFunctions[0].Percentage > 50 {
		text.WriteString("\n⚠️  High percentage of [unknown] symbols detected!\n")
		text.WriteString("\nPossible causes:\n")
		text.WriteString("  • Binary is stripped (compiled without debug symbols)\n")
		text.WriteString("  • Missing debug packages\n")
		text.WriteString("  • Compiler optimizations (inlined functions)\n")
		text.WriteString("\nRecommendations:\n")
		text.WriteString("  1. Install debug symbols for the process:\n")
		text.WriteString("     Ubuntu/Debian: apt install <package>-dbg or <package>-dbgsym\n")
		text.WriteString("     RHEL/CentOS:   yum install <package>-debuginfo\n")
		text.WriteString("  2. Check if binary is stripped: file /path/to/binary\n")
		text.WriteString("  3. For ScyllaDB: Install scylla-debuginfo package\n")
		text.WriteString("  4. Recompile with -g flag if source is available\n")
	}

	return text.String()
}
