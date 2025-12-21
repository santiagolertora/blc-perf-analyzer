package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/santiagolertora/blc-perf-analyzer/internal/analysis"
	"github.com/santiagolertora/blc-perf-analyzer/internal/capture"
	"github.com/santiagolertora/blc-perf-analyzer/internal/detector"
	"github.com/spf13/cobra"
)

var (
	// Build information
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"

	// Flags
	processName        string
	pid                int
	duration           int
	generateFlamegraph bool
	generateHeatmap    bool
	heatmapWindowSize  float64
	showVersion        bool
)

var rootCmd = &cobra.Command{
	Use:   "blc-perf-analyzer",
	Short: "Automated CPU trace analysis tool for Linux (perf)",
	Long: `BLC Perf Analyzer - Automated CPU Performance Analysis
Author: Santiago Lertora (https://santiagolertora.com)

An open source tool that automates the capture and analysis of CPU traces 
using perf, designed to detect and analyze bottlenecks in Linux processes.

When to use it?
- Troubleshooting high CPU usage in production or staging environments
- Performance tuning of databases (e.g., MariaDB), application servers, or any Linux process
- Quickly identifying userland vs. kernel bottlenecks
- Generating flamegraphs for visualization and reporting
- When you want actionable summaries without manual perf scripting

Target users: SREs, DBAs, performance engineers, DevOps, and anyone needing 
to understand process internals under load.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Detectar sistema y verificar requisitos
		sysInfo, err := detector.DetectSystem()
		if err != nil {
			return fmt.Errorf("error detecting system: %v", err)
		}

		if !sysInfo.PerfInstalled {
			fmt.Printf("perf is not installed. Attempting to install on %s...\n", sysInfo.Distro)
			if err := detector.InstallPerf(sysInfo.Distro); err != nil {
				return fmt.Errorf("error installing perf: %v", err)
			}
		}

		// 2. Verificar permisos
		if err := detector.CheckPermissions(); err != nil {
			return fmt.Errorf("error checking permissions: %v", err)
		}

		// 3. Preparar directorio de salida
		timestamp := time.Now().Format("20060102-150405")
		outputDir := filepath.Join(".", fmt.Sprintf("blc-perf-analyzer-%s", timestamp))

		// 4. Configurar y ejecutar captura
		config := &capture.CaptureConfig{
			ProcessName: processName,
			PID:         pid,
			Duration:    duration,
			OutputDir:   outputDir,
		}

		result, err := capture.Capture(config)
		if err != nil {
			return fmt.Errorf("error during capture: %v", err)
		}

		// 5. Procesar resultados y generar reportes
		if generateFlamegraph || generateHeatmap {
			fmt.Println("Generating analysis reports...")
			if err := analysis.GenerateReport(result.PerfDataPath, outputDir, processName, pid, duration, generateHeatmap, heatmapWindowSize); err != nil {
				return fmt.Errorf("error generating reports: %v", err)
			}
		} else {
			// Solo procesar perf script si no se genera flamegraph ni heatmap
			if err := capture.ProcessCapture(result); err != nil {
				return fmt.Errorf("error processing capture: %v", err)
			}
		}

		fmt.Printf("\n✓ Analysis complete. Results saved in: %s\n", outputDir)
		fmt.Println("\n📁 Generated files:")
		fmt.Println("   - perf.data: Raw perf data")

		if generateFlamegraph || generateHeatmap {
			fmt.Println("   - summary.json: Detailed analysis in JSON format")
			fmt.Println("   - summary.txt: Human-readable analysis summary")
			fmt.Println("   - perf-report.txt: Detailed perf report")
		}

		if generateFlamegraph {
			fmt.Println("   - flamegraph.svg: Interactive flamegraph visualization")
			fmt.Println("   - perf.folded: Folded stack traces")
		}

		if generateHeatmap {
			fmt.Println("   - heatmap.html: Interactive temporal heatmap")
			fmt.Println("   - heatmap-data.json: Heatmap data in JSON format")
			fmt.Println("   - patterns.json: Detected performance patterns and anomalies")
		}

		if !generateFlamegraph && !generateHeatmap {
			fmt.Println("   - perf-output.txt: Processed perf script output")
		}

		fmt.Println("\n💡 Tips:")
		fmt.Println("   - Use --generate-flamegraph to visualize call stacks")
		fmt.Println("   - Use --generate-heatmap to see performance over time")
		fmt.Println("   - Combine both flags for comprehensive analysis")

		return nil
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&processName, "process", "p", "", "Name of the process to analyze (e.g., 'mariadbd', 'nginx')")
	rootCmd.PersistentFlags().IntVar(&pid, "pid", 0, "PID of the process to analyze (e.g., 1234)")
	rootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 30, "Capture duration in seconds (default: 30)")
	rootCmd.PersistentFlags().BoolVar(&generateFlamegraph, "generate-flamegraph", false, "Generate a flamegraph SVG visualization")
	rootCmd.PersistentFlags().BoolVar(&generateHeatmap, "generate-heatmap", false, "Generate an interactive temporal heatmap")
	rootCmd.PersistentFlags().Float64Var(&heatmapWindowSize, "heatmap-window-size", 1.0, "Time window size in seconds for heatmap (default: 1.0)")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version information")

	// Validation
	rootCmd.MarkFlagsMutuallyExclusive("process", "pid")

	// Add custom validation
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Handle version flag
		if showVersion {
			printVersion()
			os.Exit(0)
		}

		if processName == "" && pid == 0 {
			return fmt.Errorf("either --process or --pid must be specified")
		}
		if processName != "" {
			// Check if process name looks like a number
			if _, err := strconv.Atoi(processName); err == nil {
				return fmt.Errorf("--process flag expects a process name (e.g., 'mariadbd'), not a number. Use --pid for process IDs")
			}
		}
		if pid != 0 && pid < 1 {
			return fmt.Errorf("PID must be a positive number")
		}
		if duration < 1 {
			return fmt.Errorf("duration must be at least 1 second")
		}
		if heatmapWindowSize <= 0 {
			return fmt.Errorf("heatmap window size must be positive")
		}
		if heatmapWindowSize > float64(duration) {
			return fmt.Errorf("heatmap window size cannot be larger than capture duration")
		}
		return nil
	}
}

func printVersion() {
	fmt.Printf("BLC Perf Analyzer %s\n", Version)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Println()
	fmt.Println("Author: Santiago Lertora")
	fmt.Println("Website: https://santiagolertora.com")
	fmt.Println("GitHub: https://github.com/santiagolertora/blc-perf-analyzer")
	fmt.Println("License: MIT")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
