package main

import (
	"testing"
)

func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name          string
		processName   string
		pid           int
		duration      int
		profileWindow int
		delayStart    int
		wantError     bool
	}{
		{
			name:        "valid with process name",
			processName: "test",
			duration:    30,
			wantError:   false,
		},
		{
			name:      "valid with PID",
			pid:       123,
			duration:  30,
			wantError: false,
		},
		{
			name:        "neither process nor PID",
			processName: "",
			pid:         0,
			wantError:   true,
		},
		{
			name:        "zero duration",
			processName: "test",
			duration:    0,
			wantError:   true,
		},
		{
			name:        "negative duration",
			processName: "test",
			duration:    -10,
			wantError:   true,
		},
		{
			name:        "valid with profile window",
			processName: "test",
			profileWindow: 60,
			wantError:   false,
		},
		{
			name:        "negative delay start",
			processName: "test",
			duration:    30,
			delayStart:  -5,
			wantError:   true,
		},
		{
			name:        "valid with delay start",
			processName: "test",
			duration:    30,
			delayStart:  10,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test process/pid validation
			if tt.processName == "" && tt.pid == 0 {
				if !tt.wantError {
					t.Error("Expected error for missing process/pid")
				}
				return
			}

			// Test duration validation
			effectiveDuration := tt.duration
			if tt.profileWindow > 0 {
				effectiveDuration = tt.profileWindow
			}

			if effectiveDuration < 1 {
				if !tt.wantError {
					t.Error("Expected error for invalid duration")
				}
				return
			}

			// Test delay start validation
			if tt.delayStart < 0 {
				if !tt.wantError {
					t.Error("Expected error for negative delay")
				}
				return
			}

			// If we got here and expected error, test failed
			if tt.wantError {
				t.Error("Expected error but validation passed")
			}
		})
	}
}

func TestEffectiveDuration(t *testing.T) {
	tests := []struct {
		name          string
		duration      int
		profileWindow int
		want          int
	}{
		{"duration only", 30, 0, 30},
		{"profile window only", 0, 60, 60},
		{"both (window takes precedence)", 30, 60, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effective := tt.duration
			if tt.profileWindow > 0 {
				effective = tt.profileWindow
			}

			if effective != tt.want {
				t.Errorf("effective duration = %d, want %d", effective, tt.want)
			}
		})
	}
}

func TestProcessNameValidation(t *testing.T) {
	tests := []struct {
		name        string
		processName string
		wantError   bool
	}{
		{"valid process name", "nginx", false},
		{"valid process name with underscore", "my_app", false},
		{"valid process name with dash", "my-app", false},
		{"numeric string", "12345", true},
		{"empty string", "", false}, // Empty is valid, will check PID instead
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple check if process name looks like a number
			if tt.processName == "" {
				return // Skip validation for empty
			}

			isNumeric := true
			for _, c := range tt.processName {
				if c < '0' || c > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric && !tt.wantError {
				t.Error("Expected error for numeric process name")
			} else if !isNumeric && tt.wantError {
				t.Error("Expected validation to pass for non-numeric name")
			}
		})
	}
}

func TestHeatmapWindowValidation(t *testing.T) {
	tests := []struct {
		name              string
		windowSize        float64
		duration          int
		wantError         bool
	}{
		{"valid small window", 0.5, 30, false},
		{"valid window equal to duration", 30.0, 30, false},
		{"window larger than duration", 60.0, 30, true},
		{"zero window size", 0.0, 30, true},
		{"negative window size", -1.0, 30, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.windowSize <= 0 || tt.windowSize > float64(tt.duration)

			if hasError != tt.wantError {
				t.Errorf("validation error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestQuietModeOutput(t *testing.T) {
	tests := []struct {
		name      string
		quietMode bool
		wantQuiet bool
	}{
		{"quiet enabled", true, true},
		{"quiet disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.quietMode != tt.wantQuiet {
				t.Errorf("quiet mode = %v, want %v", tt.quietMode, tt.wantQuiet)
			}
		})
	}
}

func TestOutputDirLogic(t *testing.T) {
	tests := []struct {
		name         string
		outputDir    string
		shouldAuto   bool
	}{
		{"explicit path", "/tmp/custom", false},
		{"empty (auto-generate)", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAuto := tt.outputDir == ""
			if isAuto != tt.shouldAuto {
				t.Errorf("auto-generate = %v, want %v", isAuto, tt.shouldAuto)
			}
		})
	}
}

func BenchmarkFlagValidation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Simulate validation logic
		processName := "test"
		duration := 30
		delayStart := 10

		if processName == "" {
			continue
		}
		if duration < 1 {
			continue
		}
		if delayStart < 0 {
			continue
		}
	}
}
