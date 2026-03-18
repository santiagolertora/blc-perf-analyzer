package capture

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/santiagolertora/blc-perf-analyzer/internal/process"
)

// CaptureConfig contains the configuration for the capture
type CaptureConfig struct {
	ProcessName string
	PID         int
	Duration    int
	DelayStart  int
	OutputDir   string
	QuietMode   bool
}

// CaptureResult contains the results of the capture
type CaptureResult struct {
	PerfDataPath string
	OutputDir    string
	StartTime    time.Time
	EndTime      time.Time
	Error        error
}

// Capture executes perf capture according to the configuration
func Capture(config *CaptureConfig) (*CaptureResult, error) {
	result := &CaptureResult{
		StartTime: time.Now(),
		OutputDir: config.OutputDir,
	}

	// Validate configuration
	if config.Duration <= 0 {
		return nil, fmt.Errorf("duration must be greater than 0")
	}

	var targetPID int

	if config.PID > 0 {
		targetPID = config.PID
		// Verify that the process exists
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", config.PID)); err != nil {
			return nil, fmt.Errorf("process with PID %d does not exist: %v", config.PID, err)
		}
	} else if config.ProcessName != "" {
		// Lookup PID by process name
		pid, err := process.GetPidByName(config.ProcessName)
		if err != nil {
			return nil, fmt.Errorf("could not find PID for process '%s': %v", config.ProcessName, err)
		}
		targetPID = pid
		if !config.QuietMode {
			fmt.Printf("Found process '%s' with PID: %d\n", config.ProcessName, targetPID)
		}
	} else {
		return nil, fmt.Errorf("either PID or process name must be provided")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %v", err)
	}

	// Handle delay start
	if config.DelayStart > 0 {
		if !config.QuietMode {
			fmt.Printf("Waiting %d seconds before starting capture...\n", config.DelayStart)
		}

		// Wait with periodic process liveness checks
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		elapsed := 0
		for range ticker.C {
			elapsed++

			// Check if process is still alive
			if _, err := os.Stat(fmt.Sprintf("/proc/%d", targetPID)); err != nil {
				return nil, fmt.Errorf("process terminated during delay period (after %d seconds)", elapsed)
			}

			if !config.QuietMode && elapsed%5 == 0 {
				fmt.Printf("  ... %d/%d seconds elapsed\n", elapsed, config.DelayStart)
			}

			if elapsed >= config.DelayStart {
				break
			}
		}

		if !config.QuietMode {
			fmt.Println("Starting capture now...")
		}
	}

	// Final liveness check before capture
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", targetPID)); err != nil {
		return nil, fmt.Errorf("process with PID %d no longer exists: %v", targetPID, err)
	}

	// Build perf command
	args := []string{"record", "-g", "-p", strconv.Itoa(targetPID), "--", "sleep", strconv.Itoa(config.Duration)}

	if !config.QuietMode {
		fmt.Printf("Capturing CPU profile for %d seconds (PID: %d)...\n", config.Duration, targetPID)
	}

	// Run perf
	stderr := make([]byte, 0)

	// Add timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Duration+5)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "perf", args...)
	cmd.Dir = config.OutputDir
	cmd.Stderr = &stderrWriter{buf: &stderr}

	if err := cmd.Run(); err != nil {
		errMsg := string(stderr)
		if errMsg == "" {
			errMsg = err.Error()
		}
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("perf command timed out after %d seconds", config.Duration+5)
		}

		// Check if it's just warnings (perf.data was still generated)
		perfDataPath := filepath.Join(config.OutputDir, "perf.data")
		if _, statErr := os.Stat(perfDataPath); statErr == nil {
			// perf.data exists, so warnings are non-fatal
			if !config.QuietMode {
				fmt.Printf("Warning: perf had warnings but capture succeeded:\n%s\n", errMsg)
			}
			result.PerfDataPath = perfDataPath
			result.EndTime = time.Now()
			return result, nil
		}

		// Real error - perf.data was not generated
		result.Error = fmt.Errorf("error running perf: %s", errMsg)
		return result, result.Error
	}

	// Check that perf.data was generated
	perfDataPath := filepath.Join(config.OutputDir, "perf.data")
	if _, err := os.Stat(perfDataPath); err != nil {
		result.Error = fmt.Errorf("perf.data file not found: %v", err)
		return result, result.Error
	}

	result.PerfDataPath = perfDataPath
	result.EndTime = time.Now()

	if !config.QuietMode {
		fmt.Printf("Capture completed successfully.\n")
	}

	return result, nil
}

// stderrWriter is a helper to capture stderr output
type stderrWriter struct {
	buf *[]byte
}

func (w *stderrWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// ProcessCapture processes the captured data
func ProcessCapture(result *CaptureResult) error {
	if result.Error != nil {
		return result.Error
	}

	// Run perf script to process the data
	cmd := exec.Command("perf", "script", "-i", result.PerfDataPath)
	outputPath := filepath.Join(result.OutputDir, "perf-output.txt")

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error processing perf data: %v", err)
	}

	// Save output to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("error saving perf output: %v", err)
	}

	return nil
}
