package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Sample represents a single perf sample
type Sample struct {
	Command   string
	PID       int
	TID       int
	CPU       int
	Timestamp float64
	Event     string
	Stack     []StackFrame
}

// StackFrame represents a single frame in a call stack
type StackFrame struct {
	Address    string
	Symbol     string
	Module     string
	Offset     string
	Type       FrameType
	IsKernel   bool
	IsUserland bool
}

// FrameType categorizes the frame
type FrameType string

const (
	FrameTypeKernelCore   FrameType = "kernel_core"
	FrameTypeKernelDriver FrameType = "kernel_driver"
	FrameTypeLibC         FrameType = "libc"
	FrameTypeLibPthread   FrameType = "libpthread"
	FrameTypeLibMySQL     FrameType = "libmysql"
	FrameTypeApplication  FrameType = "application"
	FrameTypeUnknown      FrameType = "unknown"
)

// ParsePerfScript parses the output of `perf script`
func ParsePerfScript(content string) ([]*Sample, error) {
	samples := make([]*Sample, 0)
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	// Regex patterns for perf script output
	// Format 1: mysqld 12345/12346 [001] 123456.789012:     999999 cpu-clock:
	headerRegex1 := regexp.MustCompile(`^\s*(\S+)\s+(\d+)/(\d+)\s+\[(\d+)\]\s+(\d+\.\d+):\s+\d+\s+(\S+):`)
	
	// Format 2: reactor-4    3202 88019.498348:     124999 cycles:P:
	headerRegex2 := regexp.MustCompile(`^\s*(\S+)\s+(\d+)\s+(\d+\.\d+):\s+\d+\s+(\S+):`)
	
	// Stack frame patterns:
	// 	    7ffff7a0d000 __pthread_mutex_lock+0x0 (/lib/x86_64-linux-gnu/libpthread-2.31.so)
	// 	    ffffffff81234567 do_syscall_64+0x57 ([kernel.kallsyms])
	stackRegex := regexp.MustCompile(`^\s+([0-9a-fA-F]+)\s+([^\+\(]+)(?:\+0x([0-9a-fA-F]+))?\s+\(([^\)]+)\)`)
	
	var currentSample *Sample
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Try format 1 first (with TID and CPU)
		if matches := headerRegex1.FindStringSubmatch(line); matches != nil {
			// Save previous sample if exists
			if currentSample != nil {
				samples = append(samples, currentSample)
			}
			
			// Parse new sample header
			pid, _ := strconv.Atoi(matches[2])
			tid, _ := strconv.Atoi(matches[3])
			cpu, _ := strconv.Atoi(matches[4])
			timestamp, _ := strconv.ParseFloat(matches[5], 64)
			
			currentSample = &Sample{
				Command:   strings.TrimSpace(matches[1]),
				PID:       pid,
				TID:       tid,
				CPU:       cpu,
				Timestamp: timestamp,
				Event:     strings.TrimSpace(matches[6]),
				Stack:     make([]StackFrame, 0),
			}
			continue
		}
		
		// Try format 2 (without TID/CPU in header)
		if matches := headerRegex2.FindStringSubmatch(line); matches != nil {
			// Save previous sample if exists
			if currentSample != nil {
				samples = append(samples, currentSample)
			}
			
			// Parse new sample header
			pid, _ := strconv.Atoi(matches[2])
			timestamp, _ := strconv.ParseFloat(matches[3], 64)
			
			currentSample = &Sample{
				Command:   strings.TrimSpace(matches[1]),
				PID:       pid,
				TID:       pid, // Use PID as TID when not available
				CPU:       0,   // Unknown CPU
				Timestamp: timestamp,
				Event:     strings.TrimSpace(matches[4]),
				Stack:     make([]StackFrame, 0),
			}
			continue
		}
		
		// Check if this is a stack frame line
		if currentSample != nil && strings.HasPrefix(line, "\t") {
			if matches := stackRegex.FindStringSubmatch(line); matches != nil {
				frame := StackFrame{
					Address: matches[1],
					Symbol:  strings.TrimSpace(matches[2]),
					Offset:  matches[3],
					Module:  strings.TrimSpace(matches[4]),
				}
				
				// Classify the frame
				frame.Type, frame.IsKernel, frame.IsUserland = ClassifyFrame(&frame)
				
				currentSample.Stack = append(currentSample.Stack, frame)
			}
		}
	}
	
	// Don't forget the last sample
	if currentSample != nil {
		samples = append(samples, currentSample)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning perf script output: %v", err)
	}
	
	return samples, nil
}

// ClassifyFrame determines the type and category of a stack frame
func ClassifyFrame(frame *StackFrame) (FrameType, bool, bool) {
	module := strings.ToLower(frame.Module)
	symbol := strings.ToLower(frame.Symbol)
	
	// Kernel detection
	if strings.Contains(module, "kernel.kallsyms") || 
	   strings.Contains(module, "[kernel") ||
	   strings.Contains(module, "vmlinux") {
		return FrameTypeKernelCore, true, false
	}
	
	// Kernel modules/drivers
	if strings.HasPrefix(module, "[") && strings.HasSuffix(module, "]") {
		// Could be kernel module
		return FrameTypeKernelDriver, true, false
	}
	
	// LibC
	if strings.Contains(module, "libc") && 
	   (strings.Contains(module, ".so") || strings.Contains(module, "libc-")) {
		return FrameTypeLibC, false, true
	}
	
	// LibPthread
	if strings.Contains(module, "libpthread") {
		return FrameTypeLibPthread, false, true
	}
	
	// MySQL/MariaDB libraries
	if strings.Contains(module, "mysql") || 
	   strings.Contains(module, "mariadb") ||
	   strings.Contains(symbol, "mysql") ||
	   strings.Contains(symbol, "maria") {
		return FrameTypeLibMySQL, false, true
	}
	
	// Application binary (not a shared library)
	if !strings.Contains(module, ".so") && !strings.HasPrefix(module, "[") {
		return FrameTypeApplication, false, true
	}
	
	// Default: userland unknown
	if strings.Contains(module, ".so") {
		return FrameTypeUnknown, false, true
	}
	
	return FrameTypeUnknown, false, false
}

// GetTopFrame returns the top frame of the stack (leaf function)
func (s *Sample) GetTopFrame() *StackFrame {
	if len(s.Stack) > 0 {
		return &s.Stack[0]
	}
	return nil
}

// GetBottomFrame returns the bottom frame of the stack (root)
func (s *Sample) GetBottomFrame() *StackFrame {
	if len(s.Stack) > 0 {
		return &s.Stack[len(s.Stack)-1]
	}
	return nil
}

// GetFullStack returns the full stack as a semicolon-separated string
func (s *Sample) GetFullStack() string {
	frames := make([]string, len(s.Stack))
	for i, frame := range s.Stack {
		frames[i] = frame.Symbol
	}
	return strings.Join(frames, ";")
}

// TimeWindow represents a time bucket for temporal analysis
type TimeWindow struct {
	StartTime float64
	EndTime   float64
	Duration  float64
	Samples   []*Sample
}

// PartitionByTime divides samples into time windows
func PartitionByTime(samples []*Sample, windowSizeSeconds float64) []*TimeWindow {
	if len(samples) == 0 {
		return []*TimeWindow{}
	}
	
	// Find min and max timestamps
	minTime := samples[0].Timestamp
	maxTime := samples[0].Timestamp
	
	for _, sample := range samples {
		if sample.Timestamp < minTime {
			minTime = sample.Timestamp
		}
		if sample.Timestamp > maxTime {
			maxTime = sample.Timestamp
		}
	}
	
	// Calculate number of windows needed
	totalDuration := maxTime - minTime
	numWindows := int(totalDuration/windowSizeSeconds) + 1
	
	windows := make([]*TimeWindow, numWindows)
	for i := 0; i < numWindows; i++ {
		startTime := minTime + float64(i)*windowSizeSeconds
		endTime := startTime + windowSizeSeconds
		windows[i] = &TimeWindow{
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  windowSizeSeconds,
			Samples:   make([]*Sample, 0),
		}
	}
	
	// Assign samples to windows
	for _, sample := range samples {
		windowIndex := int((sample.Timestamp - minTime) / windowSizeSeconds)
		if windowIndex >= 0 && windowIndex < numWindows {
			windows[windowIndex].Samples = append(windows[windowIndex].Samples, sample)
		}
	}
	
	return windows
}

// GetRelativeTime returns the time relative to the first sample
func (tw *TimeWindow) GetRelativeTime(firstSampleTime float64) time.Duration {
	return time.Duration((tw.StartTime - firstSampleTime) * float64(time.Second))
}

// GetTopFunctions returns the top N functions in this time window
func (tw *TimeWindow) GetTopFunctions(n int) map[string]int {
	functionCounts := make(map[string]int)
	
	for _, sample := range tw.Samples {
		if frame := sample.GetTopFrame(); frame != nil {
			functionCounts[frame.Symbol]++
		}
	}
	
	return functionCounts
}

// GetCategoryDistribution returns the distribution of frame types
func (tw *TimeWindow) GetCategoryDistribution() map[FrameType]int {
	distribution := make(map[FrameType]int)
	
	for _, sample := range tw.Samples {
		if frame := sample.GetTopFrame(); frame != nil {
			distribution[frame.Type]++
		}
	}
	
	return distribution
}

