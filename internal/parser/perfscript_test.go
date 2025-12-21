package parser

import (
	"strings"
	"testing"
)

func TestParsePerfScript(t *testing.T) {
	testInput := `mysqld 12345/12346 [001] 123456.789012:     999999 cpu-clock: 
	    7ffff7a0d000 __pthread_mutex_lock+0x0 (/lib/x86_64-linux-gnu/libpthread-2.31.so)
	    55555560abcd handle_connection+0x123 (/usr/sbin/mysqld)
	    ffffffff81234567 do_syscall_64+0x57 ([kernel.kallsyms])

mysqld 12345/12347 [002] 123456.890123:     999999 cpu-clock: 
	    7ffff7b0e111 malloc+0x45 (/lib/x86_64-linux-gnu/libc-2.31.so)
	    55555560deed query_handler+0x89 (/usr/sbin/mysqld)

mariadbd 23456/23457 [003] 123457.000000:     999999 cpu-clock: 
	    ffffffff82345678 __schedule+0x100 ([kernel.kallsyms])
	    7ffff7c0f222 pthread_cond_wait+0x12 (/lib/x86_64-linux-gnu/libpthread-2.31.so)
`

	samples, err := ParsePerfScript(testInput)
	if err != nil {
		t.Fatalf("ParsePerfScript failed: %v", err)
	}

	if len(samples) != 3 {
		t.Errorf("Expected 3 samples, got %d", len(samples))
	}

	// Test first sample
	if samples[0].Command != "mysqld" {
		t.Errorf("Expected command 'mysqld', got '%s'", samples[0].Command)
	}
	if samples[0].PID != 12345 {
		t.Errorf("Expected PID 12345, got %d", samples[0].PID)
	}
	if samples[0].TID != 12346 {
		t.Errorf("Expected TID 12346, got %d", samples[0].TID)
	}
	if samples[0].CPU != 1 {
		t.Errorf("Expected CPU 1, got %d", samples[0].CPU)
	}
	if len(samples[0].Stack) != 3 {
		t.Errorf("Expected 3 stack frames, got %d", len(samples[0].Stack))
	}

	// Test stack frame classification
	frame := samples[0].Stack[0]
	if frame.Symbol != "__pthread_mutex_lock" {
		t.Errorf("Expected symbol '__pthread_mutex_lock', got '%s'", frame.Symbol)
	}
	if frame.Type != FrameTypeLibPthread {
		t.Errorf("Expected FrameTypeLibPthread, got %s", frame.Type)
	}
	if frame.IsKernel {
		t.Error("Expected IsKernel to be false for pthread function")
	}
	if !frame.IsUserland {
		t.Error("Expected IsUserland to be true for pthread function")
	}

	// Test kernel frame
	kernelFrame := samples[0].Stack[2]
	if !kernelFrame.IsKernel {
		t.Error("Expected IsKernel to be true for kernel.kallsyms function")
	}
	if kernelFrame.Type != FrameTypeKernelCore {
		t.Errorf("Expected FrameTypeKernelCore, got %s", kernelFrame.Type)
	}
}

func TestParsePerfScriptFormat2(t *testing.T) {
	// Test format without TID/CPU (like ScyllaDB output)
	testInput := `reactor-4    3202 88019.498348:     124999 cycles:P: 
	         1caa86e [unknown] (/opt/scylladb/libexec/scylla)

reactor-5    3204 88019.501997:     124999 cycles:P: 
	ffffffffb020121a __irqentry_text_end+0xca ([kernel.kallsyms])
	         197d9af [unknown] (/opt/scylladb/libexec/scylla)
`

	samples, err := ParsePerfScript(testInput)
	if err != nil {
		t.Fatalf("ParsePerfScript failed: %v", err)
	}

	if len(samples) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(samples))
	}

	// Test first sample
	if samples[0].Command != "reactor-4" {
		t.Errorf("Expected command 'reactor-4', got '%s'", samples[0].Command)
	}
	if samples[0].PID != 3202 {
		t.Errorf("Expected PID 3202, got %d", samples[0].PID)
	}
	if len(samples[0].Stack) != 1 {
		t.Errorf("Expected 1 stack frame, got %d", len(samples[0].Stack))
	}

	// Test second sample with kernel frame
	if len(samples[1].Stack) != 2 {
		t.Errorf("Expected 2 stack frames in second sample, got %d", len(samples[1].Stack))
	}
	
	kernelFrame := samples[1].Stack[0]
	if !kernelFrame.IsKernel {
		t.Error("Expected IsKernel to be true for kernel.kallsyms function")
	}
}

func TestClassifyFrame(t *testing.T) {
	tests := []struct {
		name           string
		frame          StackFrame
		expectedType   FrameType
		expectedKernel bool
		expectedUser   bool
	}{
		{
			name:           "Kernel kallsyms",
			frame:          StackFrame{Symbol: "do_syscall_64", Module: "[kernel.kallsyms]"},
			expectedType:   FrameTypeKernelCore,
			expectedKernel: true,
			expectedUser:   false,
		},
		{
			name:           "LibPthread",
			frame:          StackFrame{Symbol: "pthread_mutex_lock", Module: "/lib/x86_64-linux-gnu/libpthread-2.31.so"},
			expectedType:   FrameTypeLibPthread,
			expectedKernel: false,
			expectedUser:   true,
		},
		{
			name:           "LibC",
			frame:          StackFrame{Symbol: "malloc", Module: "/lib/x86_64-linux-gnu/libc-2.31.so"},
			expectedType:   FrameTypeLibC,
			expectedKernel: false,
			expectedUser:   true,
		},
		{
			name:           "MySQL Library by module",
			frame:          StackFrame{Symbol: "execute_query", Module: "/usr/lib/libmysqlclient.so.21"},
			expectedType:   FrameTypeLibMySQL,
			expectedKernel: false,
			expectedUser:   true,
		},
		{
			name:           "MySQL Library by symbol",
			frame:          StackFrame{Symbol: "mysql_query", Module: "/usr/lib/some.so"},
			expectedType:   FrameTypeLibMySQL,
			expectedKernel: false,
			expectedUser:   true,
		},
		{
			name:           "Application Binary",
			frame:          StackFrame{Symbol: "main", Module: "/usr/sbin/myapp"},
			expectedType:   FrameTypeApplication,
			expectedKernel: false,
			expectedUser:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotKernel, gotUser := ClassifyFrame(&tt.frame)
			if gotType != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, gotType)
			}
			if gotKernel != tt.expectedKernel {
				t.Errorf("Expected IsKernel=%v, got %v", tt.expectedKernel, gotKernel)
			}
			if gotUser != tt.expectedUser {
				t.Errorf("Expected IsUserland=%v, got %v", tt.expectedUser, gotUser)
			}
		})
	}
}

func TestPartitionByTime(t *testing.T) {
	samples := []*Sample{
		{Timestamp: 100.0, Command: "test", PID: 1, TID: 1},
		{Timestamp: 100.5, Command: "test", PID: 1, TID: 1},
		{Timestamp: 101.0, Command: "test", PID: 1, TID: 1},
		{Timestamp: 101.5, Command: "test", PID: 1, TID: 1},
		{Timestamp: 102.0, Command: "test", PID: 1, TID: 1},
		{Timestamp: 103.0, Command: "test", PID: 1, TID: 1},
	}

	windows := PartitionByTime(samples, 1.0)

	if len(windows) < 3 {
		t.Errorf("Expected at least 3 windows for 3+ seconds of data, got %d", len(windows))
	}

	// Check that samples are distributed correctly
	totalSamplesInWindows := 0
	for _, window := range windows {
		totalSamplesInWindows += len(window.Samples)
	}

	if totalSamplesInWindows != len(samples) {
		t.Errorf("Expected %d total samples in windows, got %d", len(samples), totalSamplesInWindows)
	}

	// Check window boundaries
	for i, window := range windows {
		if window.EndTime <= window.StartTime {
			t.Errorf("Window %d has invalid time range: start=%f, end=%f", i, window.StartTime, window.EndTime)
		}
		for _, sample := range window.Samples {
			if sample.Timestamp < window.StartTime || sample.Timestamp > window.EndTime {
				t.Errorf("Sample with timestamp %f is outside window [%f, %f]", 
					sample.Timestamp, window.StartTime, window.EndTime)
			}
		}
	}
}

func TestSampleMethods(t *testing.T) {
	sample := &Sample{
		Stack: []StackFrame{
			{Symbol: "leaf_function", Type: FrameTypeApplication},
			{Symbol: "middle_function", Type: FrameTypeLibC},
			{Symbol: "root_function", Type: FrameTypeKernelCore},
		},
	}

	// Test GetTopFrame (should return first/leaf)
	topFrame := sample.GetTopFrame()
	if topFrame == nil {
		t.Fatal("GetTopFrame returned nil")
	}
	if topFrame.Symbol != "leaf_function" {
		t.Errorf("Expected top frame 'leaf_function', got '%s'", topFrame.Symbol)
	}

	// Test GetBottomFrame (should return last/root)
	bottomFrame := sample.GetBottomFrame()
	if bottomFrame == nil {
		t.Fatal("GetBottomFrame returned nil")
	}
	if bottomFrame.Symbol != "root_function" {
		t.Errorf("Expected bottom frame 'root_function', got '%s'", bottomFrame.Symbol)
	}

	// Test GetFullStack
	fullStack := sample.GetFullStack()
	expected := "leaf_function;middle_function;root_function"
	if fullStack != expected {
		t.Errorf("Expected full stack '%s', got '%s'", expected, fullStack)
	}
}

func TestTimeWindowGetTopFunctions(t *testing.T) {
	samples := []*Sample{
		{
			Stack: []StackFrame{
				{Symbol: "function_a", Type: FrameTypeApplication},
			},
		},
		{
			Stack: []StackFrame{
				{Symbol: "function_a", Type: FrameTypeApplication},
			},
		},
		{
			Stack: []StackFrame{
				{Symbol: "function_b", Type: FrameTypeLibC},
			},
		},
	}

	window := &TimeWindow{
		Samples: samples,
	}

	topFunctions := window.GetTopFunctions(10)

	if len(topFunctions) != 2 {
		t.Errorf("Expected 2 unique functions, got %d", len(topFunctions))
	}

	if topFunctions["function_a"] != 2 {
		t.Errorf("Expected function_a count 2, got %d", topFunctions["function_a"])
	}

	if topFunctions["function_b"] != 1 {
		t.Errorf("Expected function_b count 1, got %d", topFunctions["function_b"])
	}
}

func TestTimeWindowGetCategoryDistribution(t *testing.T) {
	samples := []*Sample{
		{Stack: []StackFrame{{Type: FrameTypeKernelCore, IsKernel: true}}},
		{Stack: []StackFrame{{Type: FrameTypeKernelCore, IsKernel: true}}},
		{Stack: []StackFrame{{Type: FrameTypeLibC, IsUserland: true}}},
		{Stack: []StackFrame{{Type: FrameTypeApplication, IsUserland: true}}},
	}

	window := &TimeWindow{Samples: samples}
	distribution := window.GetCategoryDistribution()

	if distribution[FrameTypeKernelCore] != 2 {
		t.Errorf("Expected 2 kernel samples, got %d", distribution[FrameTypeKernelCore])
	}
	if distribution[FrameTypeLibC] != 1 {
		t.Errorf("Expected 1 libc sample, got %d", distribution[FrameTypeLibC])
	}
	if distribution[FrameTypeApplication] != 1 {
		t.Errorf("Expected 1 application sample, got %d", distribution[FrameTypeApplication])
	}
}

func TestParsePerfScriptEmptyInput(t *testing.T) {
	samples, err := ParsePerfScript("")
	if err != nil {
		t.Fatalf("ParsePerfScript with empty input should not error: %v", err)
	}
	if len(samples) != 0 {
		t.Errorf("Expected 0 samples from empty input, got %d", len(samples))
	}
}

func TestParsePerfScriptMalformedInput(t *testing.T) {
	malformedInput := `this is not valid perf script output
random text here
12345 more random stuff`

	samples, err := ParsePerfScript(malformedInput)
	if err != nil {
		t.Fatalf("ParsePerfScript should handle malformed input gracefully: %v", err)
	}
	// Should return empty slice, not error
	if len(samples) != 0 {
		t.Errorf("Expected 0 samples from malformed input, got %d", len(samples))
	}
}

func BenchmarkParsePerfScript(b *testing.B) {
	// Create a realistic perf script output
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("mysqld 12345/12346 [001] 123456.789012:     999999 cpu-clock:\n")
		sb.WriteString("\t    7ffff7a0d000 __pthread_mutex_lock+0x0 (/lib/x86_64-linux-gnu/libpthread-2.31.so)\n")
		sb.WriteString("\t    55555560abcd handle_connection+0x123 (/usr/sbin/mysqld)\n")
		sb.WriteString("\t    ffffffff81234567 do_syscall_64+0x57 ([kernel.kallsyms])\n\n")
	}
	input := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParsePerfScript(input)
	}
}

func BenchmarkClassifyFrame(b *testing.B) {
	frame := &StackFrame{
		Symbol: "pthread_mutex_lock",
		Module: "/lib/x86_64-linux-gnu/libpthread-2.31.so",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ClassifyFrame(frame)
	}
}

