package analysis

import (
	"testing"

	"github.com/santiagolertora/blc-perf-analyzer/internal/parser"
)

func TestParsePerfReport(t *testing.T) {
	// Create test samples
	samples := []*parser.Sample{
		{
			Stack: []parser.StackFrame{
				{
					Symbol:     "pthread_mutex_lock",
					Module:     "/lib/libpthread.so",
					Type:       parser.FrameTypeLibPthread,
					IsUserland: true,
				},
			},
		},
		{
			Stack: []parser.StackFrame{
				{
					Symbol:   "do_syscall_64",
					Module:   "[kernel.kallsyms]",
					Type:     parser.FrameTypeKernelCore,
					IsKernel: true,
				},
			},
		},
		{
			Stack: []parser.StackFrame{
				{
					Symbol:     "malloc",
					Module:     "/lib/libc.so",
					Type:       parser.FrameTypeLibC,
					IsUserland: true,
				},
			},
		},
		{
			Stack: []parser.StackFrame{
				{
					Symbol:   "schedule",
					Module:   "[kernel.kallsyms]",
					Type:     parser.FrameTypeKernelCore,
					IsKernel: true,
				},
			},
		},
	}

	result := parsePerfReport("", samples)

	if result == nil {
		t.Fatal("parsePerfReport returned nil")
	}

	// Check total samples
	if result.Summary.TotalSamples != 4 {
		t.Errorf("Expected 4 total samples, got %d", result.Summary.TotalSamples)
	}

	// Check kernel percentage (2 out of 4 = 50%)
	expectedKernel := 50.0
	if result.Summary.KernelPercent != expectedKernel {
		t.Errorf("Expected kernel percent %.1f, got %.1f", expectedKernel, result.Summary.KernelPercent)
	}

	// Check userland percentage (2 out of 4 = 50%)
	expectedUserland := 50.0
	if result.Summary.UserlandPercent != expectedUserland {
		t.Errorf("Expected userland percent %.1f, got %.1f", expectedUserland, result.Summary.UserlandPercent)
	}

	// Check that we have function stats
	if len(result.TopFunctions) == 0 {
		t.Error("Expected some functions in TopFunctions, got none")
	}

	// Verify functions are sorted by total samples
	for i := 0; i < len(result.TopFunctions)-1; i++ {
		if result.TopFunctions[i].TotalSamples < result.TopFunctions[i+1].TotalSamples {
			t.Errorf("TopFunctions not sorted correctly at index %d", i)
		}
	}
}

func TestParsePerfReportEmptySamples(t *testing.T) {
	result := parsePerfReport("", []*parser.Sample{})

	if result == nil {
		t.Fatal("parsePerfReport returned nil")
	}

	if result.Summary.TotalSamples != 0 {
		t.Errorf("Expected 0 total samples, got %d", result.Summary.TotalSamples)
	}

	if len(result.TopFunctions) != 0 {
		t.Errorf("Expected 0 top functions, got %d", len(result.TopFunctions))
	}
}

func TestGenerateSummaryText(t *testing.T) {
	summary := SummaryStats{
		ProcessName:     "test_process",
		PID:             12345,
		CaptureDuration: 60,
		TotalSamples:    1000,
		UserlandPercent: 70.5,
		KernelPercent:   25.3,
		UnknownPercent:  4.2,
	}

	topFunctions := []FunctionStats{
		{Name: "function_a", Percentage: 15.5, TotalSamples: 155},
		{Name: "function_b", Percentage: 12.3, TotalSamples: 123},
		{Name: "function_c", Percentage: 10.1, TotalSamples: 101},
	}

	text := generateSummaryText(summary, topFunctions)

	// Check that text contains expected elements
	if text == "" {
		t.Fatal("generateSummaryText returned empty string")
	}

	requiredStrings := []string{
		"Performance Analysis Summary",
		"test_process",
		"12345",
		"60 seconds",
		"1000",
		"70.5",
		"25.3",
		"function_a",
		"function_b",
		"function_c",
	}

	for _, required := range requiredStrings {
		if !contains(text, required) {
			t.Errorf("Summary text missing required string: %s", required)
		}
	}
}

func TestFunctionStatsPercentageCalculation(t *testing.T) {
	samples := make([]*parser.Sample, 100)
	for i := 0; i < 100; i++ {
		var frameType parser.FrameType
		var isKernel, isUserland bool

		if i < 30 {
			// 30% function_a
			frameType = parser.FrameTypeApplication
			isUserland = true
		} else if i < 50 {
			// 20% function_b
			frameType = parser.FrameTypeLibC
			isUserland = true
		} else {
			// 50% kernel
			frameType = parser.FrameTypeKernelCore
			isKernel = true
		}

		var symbol string
		if i < 30 {
			symbol = "function_a"
		} else if i < 50 {
			symbol = "function_b"
		} else {
			symbol = "kernel_func"
		}

		samples[i] = &parser.Sample{
			Stack: []parser.StackFrame{
				{
					Symbol:     symbol,
					Type:       frameType,
					IsKernel:   isKernel,
					IsUserland: isUserland,
				},
			},
		}
	}

	result := parsePerfReport("", samples)

	// Find function_a in results
	var funcA *FunctionStats
	for i := range result.TopFunctions {
		if result.TopFunctions[i].Name == "function_a" {
			funcA = &result.TopFunctions[i]
			break
		}
	}

	if funcA == nil {
		t.Fatal("function_a not found in results")
	}

	// Should be 30%
	if funcA.Percentage != 30.0 {
		t.Errorf("Expected function_a percentage 30.0, got %.1f", funcA.Percentage)
	}

	if funcA.TotalSamples != 30 {
		t.Errorf("Expected function_a total samples 30, got %d", funcA.TotalSamples)
	}

	// Check userland percentage (50 out of 100)
	if result.Summary.UserlandPercent != 50.0 {
		t.Errorf("Expected userland percent 50.0, got %.1f", result.Summary.UserlandPercent)
	}

	// Check kernel percentage (50 out of 100)
	if result.Summary.KernelPercent != 50.0 {
		t.Errorf("Expected kernel percent 50.0, got %.1f", result.Summary.KernelPercent)
	}
}

func TestProcessPerfOutput(t *testing.T) {
	// Test the folded stack generation
	input := `process 1234 [000] 123.456: cpu-clock: 
	    7ffff7a0d000 function_a+0x10 (/lib/test.so)
	    55555560abcd function_b+0x20 (/usr/bin/app)

process 1234 [000] 124.456: cpu-clock: 
	    7ffff7a0d000 function_a+0x10 (/lib/test.so)
	    55555560abcd function_b+0x20 (/usr/bin/app)
`

	output := processPerfOutput(input)

	if output == "" {
		t.Error("processPerfOutput returned empty string")
	}

	// The output should contain stack information
	// Note: The actual format depends on the implementation
	// This is a basic check that something was generated
	if len(output) < 10 {
		t.Error("processPerfOutput generated suspiciously short output")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func BenchmarkParsePerfReport(b *testing.B) {
	// Create 1000 sample records
	samples := make([]*parser.Sample, 1000)
	for i := 0; i < 1000; i++ {
		samples[i] = &parser.Sample{
			Stack: []parser.StackFrame{
				{
					Symbol:     "test_function",
					Type:       parser.FrameTypeApplication,
					IsUserland: true,
				},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePerfReport("", samples)
	}
}

func BenchmarkGenerateSummaryText(b *testing.B) {
	summary := SummaryStats{
		ProcessName:     "test",
		PID:             12345,
		CaptureDuration: 60,
		TotalSamples:    10000,
		UserlandPercent: 65.5,
		KernelPercent:   30.2,
		UnknownPercent:  4.3,
	}

	topFunctions := make([]FunctionStats, 50)
	for i := 0; i < 50; i++ {
		topFunctions[i] = FunctionStats{
			Name:         "function_" + string(rune(i)),
			Percentage:   float64(i) / 10.0,
			TotalSamples: i * 10,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSummaryText(summary, topFunctions)
	}
}

