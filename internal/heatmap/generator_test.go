package heatmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/santiagolertora/blc-perf-analyzer/internal/parser"
)

func TestGenerateHeatmap(t *testing.T) {
	// Create test samples
	samples := createTestSamples()

	// Create temporary output directory
	tempDir := t.TempDir()

	// Generate heatmap
	err := GenerateHeatmap(samples, tempDir, "test_process", 12345, 1.0)
	if err != nil {
		t.Fatalf("GenerateHeatmap failed: %v", err)
	}

	// Verify output files exist
	expectedFiles := []string{
		"heatmap.html",
		"heatmap-data.json",
		"patterns.json",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tempDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", filename)
		}
	}

	// Verify HTML file has content
	htmlPath := filepath.Join(tempDir, "heatmap.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	if len(content) == 0 {
		t.Error("HTML file is empty")
	}

	// Verify HTML contains expected elements
	htmlStr := string(content)
	requiredElements := []string{
		"CPU Performance Heatmap",
		"test_process",
		"plotly",
		"heatmap",
		"Kernel vs Userland",
	}

	for _, elem := range requiredElements {
		if !contains(htmlStr, elem) {
			t.Errorf("HTML does not contain expected element: %s", elem)
		}
	}
}

func TestDetectPatterns(t *testing.T) {
	windows := []*TimeWindowData{
		{
			WindowIndex: 0,
			SampleCount: 100,
			FunctionCounts: map[string]int{
				"normal_function": 100,
			},
			CategoryCounts: map[string]int{
				"application": 100,
			},
			KernelPercent: 10.0,
		},
		{
			WindowIndex: 1,
			SampleCount: 100,
			FunctionCounts: map[string]int{
				"pthread_mutex_lock": 60,
				"other_function":     40,
			},
			CategoryCounts: map[string]int{
				"libpthread": 60,
			},
			KernelPercent: 10.0,
		},
		{
			WindowIndex: 2,
			SampleCount: 100,
			FunctionCounts: map[string]int{
				"normal_function": 100,
			},
			CategoryCounts: map[string]int{
				"kernel_core": 80,
			},
			KernelPercent: 80.0,
		},
		{
			WindowIndex: 3,
			SampleCount: 300, // Spike
			FunctionCounts: map[string]int{
				"hot_function": 300,
			},
			CategoryCounts: map[string]int{
				"application": 300,
			},
			KernelPercent: 10.0,
		},
	}

	patterns := detectPatterns(windows)

	// Check lock contention detection
	if len(patterns.LockContentionWindows) == 0 {
		t.Error("Expected to detect lock contention in window 1")
	}

	// Check high syscall detection
	if len(patterns.HighSyscallWindows) == 0 {
		t.Error("Expected to detect high syscall activity in window 2")
	}

	// Check CPU spike detection
	if len(patterns.CPUSpikes) == 0 {
		t.Error("Expected to detect CPU spike in window 3")
	}

	// Check anomalies
	if len(patterns.Anomalies) == 0 {
		t.Error("Expected to detect anomalies")
	}

	// Verify anomaly types
	anomalyTypes := make(map[string]bool)
	for _, anomaly := range patterns.Anomalies {
		anomalyTypes[anomaly.Type] = true
	}

	if !anomalyTypes["lock_contention"] {
		t.Error("Expected lock_contention anomaly")
	}
	if !anomalyTypes["high_syscall"] {
		t.Error("Expected high_syscall anomaly")
	}
	if !anomalyTypes["cpu_spike"] {
		t.Error("Expected cpu_spike anomaly")
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substrs  []string
		expected bool
	}{
		{
			name:     "Contains one",
			str:      "pthread_mutex_lock",
			substrs:  []string{"mutex", "futex"},
			expected: true,
		},
		{
			name:     "Contains none",
			str:      "normal_function",
			substrs:  []string{"mutex", "futex"},
			expected: false,
		},
		{
			name:     "Contains multiple",
			str:      "futex_wait",
			substrs:  []string{"mutex", "futex", "lock"},
			expected: true,
		},
		{
			name:     "Empty substrs",
			str:      "test",
			substrs:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.str, tt.substrs)
			if result != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.str, tt.substrs, result, tt.expected)
			}
		})
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substr   string
		expected bool
	}{
		{"Found at start", "hello world", "hello", true},
		{"Found at end", "hello world", "world", true},
		{"Found in middle", "hello world", "lo wo", true},
		{"Not found", "hello world", "xyz", false},
		{"Empty substr", "hello", "", true},
		{"Exact match", "test", "test", true},
		{"Substr longer", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSubstring(tt.str, tt.substr)
			if result != tt.expected {
				t.Errorf("findSubstring(%q, %q) = %v, want %v", tt.str, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestGenerateHeatmapEmptySamples(t *testing.T) {
	tempDir := t.TempDir()
	err := GenerateHeatmap([]*parser.Sample{}, tempDir, "test", 123, 1.0)
	if err == nil {
		t.Error("Expected error when generating heatmap with empty samples")
	}
}

func TestTimeWindowDataCalculations(t *testing.T) {
	// Create a window with known data
	samples := createTestSamples()
	windows := parser.PartitionByTime(samples, 1.0)

	if len(windows) == 0 {
		t.Fatal("Expected at least one window")
	}

	// Create TimeWindowData
	window := windows[0]
	twd := &TimeWindowData{
		WindowIndex:    0,
		StartTime:      window.StartTime,
		EndTime:        window.EndTime,
		SampleCount:    len(window.Samples),
		FunctionCounts: make(map[string]int),
		ThreadCounts:   make(map[int]int),
		CategoryCounts: make(map[string]int),
	}

	var kernelCount, userlandCount int
	for _, sample := range window.Samples {
		twd.ThreadCounts[sample.TID]++
		if frame := sample.GetTopFrame(); frame != nil {
			twd.FunctionCounts[frame.Symbol]++
			twd.CategoryCounts[string(frame.Type)]++
			if frame.IsKernel {
				kernelCount++
			} else if frame.IsUserland {
				userlandCount++
			}
		}
	}

	if twd.SampleCount > 0 {
		twd.KernelPercent = float64(kernelCount) / float64(twd.SampleCount) * 100
		twd.UserlandPercent = float64(userlandCount) / float64(twd.SampleCount) * 100
	}

	// Verify calculations
	if twd.KernelPercent < 0 || twd.KernelPercent > 100 {
		t.Errorf("Invalid KernelPercent: %f", twd.KernelPercent)
	}
	if twd.UserlandPercent < 0 || twd.UserlandPercent > 100 {
		t.Errorf("Invalid UserlandPercent: %f", twd.UserlandPercent)
	}

	totalPercent := twd.KernelPercent + twd.UserlandPercent
	if totalPercent > 100.1 { // Allow small floating point error
		t.Errorf("Total percent exceeds 100: %f", totalPercent)
	}
}

// Helper functions

func createTestSamples() []*parser.Sample {
	baseTime := 1000.0
	samples := make([]*parser.Sample, 0, 100)

	for i := 0; i < 100; i++ {
		sample := &parser.Sample{
			Command:   "test_process",
			PID:       12345,
			TID:       12346 + (i % 3), // 3 different threads
			CPU:       i % 4,            // 4 CPUs
			Timestamp: baseTime + float64(i)*0.1,
			Event:     "cpu-clock",
			Stack:     make([]parser.StackFrame, 0),
		}

		// Add different types of stack frames based on index
		switch i % 5 {
		case 0:
			sample.Stack = append(sample.Stack, parser.StackFrame{
				Symbol:     "pthread_mutex_lock",
				Module:     "/lib/libpthread.so",
				Type:       parser.FrameTypeLibPthread,
				IsUserland: true,
			})
		case 1:
			sample.Stack = append(sample.Stack, parser.StackFrame{
				Symbol:   "do_syscall_64",
				Module:   "[kernel.kallsyms]",
				Type:     parser.FrameTypeKernelCore,
				IsKernel: true,
			})
		case 2:
			sample.Stack = append(sample.Stack, parser.StackFrame{
				Symbol:     "malloc",
				Module:     "/lib/libc.so",
				Type:       parser.FrameTypeLibC,
				IsUserland: true,
			})
		case 3:
			sample.Stack = append(sample.Stack, parser.StackFrame{
				Symbol:     "mysql_execute",
				Module:     "/usr/lib/libmysqlclient.so",
				Type:       parser.FrameTypeLibMySQL,
				IsUserland: true,
			})
		case 4:
			sample.Stack = append(sample.Stack, parser.StackFrame{
				Symbol:     "main",
				Module:     "/usr/sbin/test_process",
				Type:       parser.FrameTypeApplication,
				IsUserland: true,
			})
		}

		samples = append(samples, sample)
	}

	return samples
}

func contains(s, substr string) bool {
	return findSubstring(s, substr)
}

func BenchmarkGenerateHeatmap(b *testing.B) {
	samples := createTestSamples()
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateHeatmap(samples, tempDir, "test", 12345, 1.0)
	}
}

func BenchmarkDetectPatterns(b *testing.B) {
	windows := make([]*TimeWindowData, 100)
	for i := 0; i < 100; i++ {
		windows[i] = &TimeWindowData{
			WindowIndex: i,
			SampleCount: 100 + (i % 50),
			FunctionCounts: map[string]int{
				"function_a":         30,
				"pthread_mutex_lock": 20,
				"futex_wait":         10,
				"normal_func":        40,
			},
			CategoryCounts: map[string]int{
				"application": 50,
				"kernel_core": 30,
				"libpthread":  20,
			},
			KernelPercent:   30.0,
			UserlandPercent: 70.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectPatterns(windows)
	}
}

