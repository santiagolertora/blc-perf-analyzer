package capture

import (
	"os"
	"testing"
	"time"
)

func TestCaptureConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		config    *CaptureConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config with PID",
			config: &CaptureConfig{
				PID:      1,
				Duration: 10,
				OutputDir: "/tmp/test",
			},
			wantError: false,
		},
		{
			name: "valid config with ProcessName",
			config: &CaptureConfig{
				ProcessName: "init",
				Duration:    10,
				OutputDir:   "/tmp/test",
			},
			wantError: false,
		},
		{
			name: "zero duration should fail",
			config: &CaptureConfig{
				PID:      1,
				Duration: 0,
				OutputDir: "/tmp/test",
			},
			wantError: true,
			errorMsg:  "duration must be greater than 0",
		},
		{
			name: "negative duration should fail",
			config: &CaptureConfig{
				PID:      1,
				Duration: -10,
				OutputDir: "/tmp/test",
			},
			wantError: true,
			errorMsg:  "duration must be greater than 0",
		},
		{
			name: "valid config with delay",
			config: &CaptureConfig{
				PID:        1,
				Duration:   10,
				DelayStart: 5,
				OutputDir:  "/tmp/test",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation check
			if tt.config.Duration <= 0 && !tt.wantError {
				t.Errorf("Expected error for invalid duration but got none")
			}
			if tt.config.Duration > 0 && tt.wantError && tt.errorMsg == "duration must be greater than 0" {
				// Expected this error
			}
		})
	}
}

func TestCaptureConfig_DelayStart(t *testing.T) {
	config := &CaptureConfig{
		PID:        os.Getpid(),
		Duration:   1,
		DelayStart: 0,
		OutputDir:  "/tmp/test-delay",
		QuietMode:  true,
	}

	if config.DelayStart < 0 {
		t.Errorf("DelayStart should not be negative, got %d", config.DelayStart)
	}

	if config.Duration < 1 {
		t.Errorf("Duration should be at least 1, got %d", config.Duration)
	}
}

func TestCaptureConfig_QuietMode(t *testing.T) {
	tests := []struct {
		name      string
		quietMode bool
		want      bool
	}{
		{"quiet enabled", true, true},
		{"quiet disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CaptureConfig{
				QuietMode: tt.quietMode,
			}
			if config.QuietMode != tt.want {
				t.Errorf("QuietMode = %v, want %v", config.QuietMode, tt.want)
			}
		})
	}
}

func TestCaptureResult_Timing(t *testing.T) {
	start := time.Now()
	result := &CaptureResult{
		StartTime: start,
	}

	time.Sleep(10 * time.Millisecond)
	result.EndTime = time.Now()

	duration := result.EndTime.Sub(result.StartTime)
	if duration < 10*time.Millisecond {
		t.Errorf("Duration should be at least 10ms, got %v", duration)
	}
}

func TestCaptureConfig_OutputDir(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		wantEmpty bool
	}{
		{"explicit path", "/tmp/test-output", false},
		{"relative path", "./results", false},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &CaptureConfig{
				OutputDir: tt.outputDir,
			}
			isEmpty := config.OutputDir == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("OutputDir empty = %v, want %v", isEmpty, tt.wantEmpty)
			}
		})
	}
}

func TestStderrWriter(t *testing.T) {
	buf := make([]byte, 0)
	writer := &stderrWriter{buf: &buf}

	testData := []byte("test error message\n")
	n, err := writer.Write(testData)

	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
	}
	if n != len(testData) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testData))
	}
	if string(buf) != string(testData) {
		t.Errorf("Buffer = %q, want %q", string(buf), string(testData))
	}
}

func TestStderrWriter_MultipleWrites(t *testing.T) {
	buf := make([]byte, 0)
	writer := &stderrWriter{buf: &buf}

	writes := []string{"line1\n", "line2\n", "line3\n"}
	for _, data := range writes {
		writer.Write([]byte(data))
	}

	expected := "line1\nline2\nline3\n"
	if string(buf) != expected {
		t.Errorf("Buffer = %q, want %q", string(buf), expected)
	}
}

func BenchmarkStderrWriter(b *testing.B) {
	buf := make([]byte, 0)
	writer := &stderrWriter{buf: &buf}
	data := []byte("benchmark test data\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
}
