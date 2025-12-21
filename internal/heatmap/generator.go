package heatmap

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/santiagolertora/blc-perf-analyzer/internal/parser"
)

// HeatmapData contains all data needed for heatmap visualization
type HeatmapData struct {
	TimeWindows      []*TimeWindowData `json:"time_windows"`
	Functions        []string          `json:"functions"`
	Threads          []int             `json:"threads"`
	WindowSize       float64           `json:"window_size_seconds"`
	TotalDuration    float64           `json:"total_duration_seconds"`
	TotalSamples     int               `json:"total_samples"`
	ProcessName      string            `json:"process_name"`
	PID              int               `json:"pid"`
	CaptureTimestamp string            `json:"capture_timestamp"`
}

// TimeWindowData represents aggregated data for a time window
type TimeWindowData struct {
	WindowIndex        int                       `json:"window_index"`
	StartTime          float64                   `json:"start_time"`
	EndTime            float64                   `json:"end_time"`
	SampleCount        int                       `json:"sample_count"`
	FunctionCounts     map[string]int            `json:"function_counts"`
	ThreadCounts       map[int]int               `json:"thread_counts"`
	CategoryCounts     map[string]int            `json:"category_counts"`
	TopFunction        string                    `json:"top_function"`
	TopFunctionPercent float64                   `json:"top_function_percent"`
	KernelPercent      float64                   `json:"kernel_percent"`
	UserlandPercent    float64                   `json:"userland_percent"`
}

// PatternDetection contains detected patterns and anomalies
type PatternDetection struct {
	LockContentionWindows []int     `json:"lock_contention_windows"`
	HighSyscallWindows    []int     `json:"high_syscall_windows"`
	CPUSpikes             []int     `json:"cpu_spikes"`
	Anomalies             []Anomaly `json:"anomalies"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	WindowIndex int     `json:"window_index"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	Value       float64 `json:"value"`
}

// GenerateHeatmap creates a comprehensive heatmap analysis
func GenerateHeatmap(samples []*parser.Sample, outputDir string, processName string, pid int, windowSize float64) error {
	if len(samples) == 0 {
		return fmt.Errorf("no samples to analyze")
	}

	// Partition samples into time windows
	windows := parser.PartitionByTime(samples, windowSize)
	
	// Extract unique functions and threads
	functionsMap := make(map[string]bool)
	threadsMap := make(map[int]bool)
	
	for _, sample := range samples {
		if frame := sample.GetTopFrame(); frame != nil {
			functionsMap[frame.Symbol] = true
		}
		threadsMap[sample.TID] = true
	}
	
	// Convert to sorted slices
	functions := make([]string, 0, len(functionsMap))
	for fn := range functionsMap {
		functions = append(functions, fn)
	}
	sort.Strings(functions)
	
	threads := make([]int, 0, len(threadsMap))
	for tid := range threadsMap {
		threads = append(threads, tid)
	}
	sort.Ints(threads)
	
	// Calculate total duration
	var totalDuration float64
	if len(windows) > 0 {
		totalDuration = windows[len(windows)-1].EndTime - windows[0].StartTime
	}
	
	// Process each time window
	timeWindowsData := make([]*TimeWindowData, len(windows))
	for i, window := range windows {
		twd := &TimeWindowData{
			WindowIndex:    i,
			StartTime:      window.StartTime,
			EndTime:        window.EndTime,
			SampleCount:    len(window.Samples),
			FunctionCounts: make(map[string]int),
			ThreadCounts:   make(map[int]int),
			CategoryCounts: make(map[string]int),
		}
		
		// Count occurrences
		var kernelCount, userlandCount int
		
		for _, sample := range window.Samples {
			// Count by thread
			twd.ThreadCounts[sample.TID]++
			
			// Count by function and category
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
		
		// Calculate percentages
		if twd.SampleCount > 0 {
			twd.KernelPercent = float64(kernelCount) / float64(twd.SampleCount) * 100
			twd.UserlandPercent = float64(userlandCount) / float64(twd.SampleCount) * 100
			
			// Find top function
			maxCount := 0
			for fn, count := range twd.FunctionCounts {
				if count > maxCount {
					maxCount = count
					twd.TopFunction = fn
				}
			}
			twd.TopFunctionPercent = float64(maxCount) / float64(twd.SampleCount) * 100
		}
		
		timeWindowsData[i] = twd
	}
	
	// Create heatmap data structure
	heatmapData := &HeatmapData{
		TimeWindows:   timeWindowsData,
		Functions:     functions,
		Threads:       threads,
		WindowSize:    windowSize,
		TotalDuration: totalDuration,
		TotalSamples:  len(samples),
		ProcessName:   processName,
		PID:           pid,
	}
	
	// Detect patterns
	patterns := detectPatterns(timeWindowsData)
	
	// Generate HTML visualization
	if err := generateHTMLHeatmap(heatmapData, patterns, outputDir); err != nil {
		return fmt.Errorf("error generating HTML heatmap: %v", err)
	}
	
	// Save JSON data
	jsonPath := filepath.Join(outputDir, "heatmap-data.json")
	jsonData, err := json.MarshalIndent(heatmapData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling heatmap data: %v", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing heatmap JSON: %v", err)
	}
	
	// Save patterns JSON
	patternsPath := filepath.Join(outputDir, "patterns.json")
	patternsData, err := json.MarshalIndent(patterns, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling patterns: %v", err)
	}
	if err := os.WriteFile(patternsPath, patternsData, 0644); err != nil {
		return fmt.Errorf("error writing patterns JSON: %v", err)
	}
	
	return nil
}

// detectPatterns analyzes time windows to detect patterns
func detectPatterns(windows []*TimeWindowData) *PatternDetection {
	patterns := &PatternDetection{
		LockContentionWindows: make([]int, 0),
		HighSyscallWindows:    make([]int, 0),
		CPUSpikes:             make([]int, 0),
		Anomalies:             make([]Anomaly, 0),
	}
	
	// Calculate average samples per window
	var totalSamples int
	for _, w := range windows {
		totalSamples += w.SampleCount
	}
	avgSamples := float64(totalSamples) / float64(len(windows))
	
	// Analyze each window
	for i, window := range windows {
		// Detect lock contention (high pthread/futex activity)
		lockCount := 0
		for fn, count := range window.FunctionCounts {
			fnLower := fmt.Sprintf("%s", fn)
			if containsAny(fnLower, []string{"pthread_mutex", "futex", "rwlock", "__lll_lock"}) {
				lockCount += count
			}
		}
		
		if lockCount > window.SampleCount/2 { // More than 50% lock-related
			patterns.LockContentionWindows = append(patterns.LockContentionWindows, i)
			patterns.Anomalies = append(patterns.Anomalies, Anomaly{
				WindowIndex: i,
				Type:        "lock_contention",
				Description: fmt.Sprintf("High lock contention detected: %d%% of samples", lockCount*100/window.SampleCount),
				Severity:    "high",
				Value:       float64(lockCount) / float64(window.SampleCount) * 100,
			})
		}
		
		// Detect high syscall activity
		syscallCount, exists := window.CategoryCounts["kernel_core"]
		if exists && syscallCount > window.SampleCount*70/100 { // More than 70% kernel
			patterns.HighSyscallWindows = append(patterns.HighSyscallWindows, i)
			patterns.Anomalies = append(patterns.Anomalies, Anomaly{
				WindowIndex: i,
				Type:        "high_syscall",
				Description: fmt.Sprintf("High kernel/syscall activity: %.1f%%", window.KernelPercent),
				Severity:    "medium",
				Value:       window.KernelPercent,
			})
		}
		
		// Detect CPU spikes (sample count significantly above average)
		if float64(window.SampleCount) > avgSamples*1.5 { // 50% above average
			patterns.CPUSpikes = append(patterns.CPUSpikes, i)
			patterns.Anomalies = append(patterns.Anomalies, Anomaly{
				WindowIndex: i,
				Type:        "cpu_spike",
				Description: fmt.Sprintf("CPU usage spike: %d samples (avg: %.0f)", window.SampleCount, avgSamples),
				Severity:    "medium",
				Value:       float64(window.SampleCount),
			})
		}
	}
	
	return patterns
}

// containsAny checks if string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && findSubstring(s, substr) {
			return true
		}
	}
	return false
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
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

// generateHTMLHeatmap creates an interactive HTML visualization
func generateHTMLHeatmap(data *HeatmapData, patterns *PatternDetection, outputDir string) error {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CPU Performance Heatmap - {{.ProcessName}}</title>
    <script src="https://cdn.plot.ly/plotly-2.26.0.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #0f0f23;
            color: #cccccc;
            padding: 20px;
        }
        .container { max-width: 1600px; margin: 0 auto; }
        h1 {
            color: #00ff00;
            text-align: center;
            margin-bottom: 10px;
            font-size: 2.5em;
            text-shadow: 0 0 10px #00ff00;
        }
        .subtitle {
            text-align: center;
            color: #888;
            margin-bottom: 30px;
            font-size: 1.1em;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #1a1a2e;
            border: 1px solid #00ff00;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 0 20px rgba(0, 255, 0, 0.2);
        }
        .stat-label {
            color: #888;
            font-size: 0.9em;
            margin-bottom: 5px;
        }
        .stat-value {
            color: #00ff00;
            font-size: 2em;
            font-weight: bold;
        }
        .chart-container {
            background: #1a1a2e;
            border: 1px solid #00ff00;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 30px;
            box-shadow: 0 0 20px rgba(0, 255, 0, 0.2);
        }
        .chart-title {
            color: #00ff00;
            font-size: 1.5em;
            margin-bottom: 15px;
            text-align: center;
        }
        .anomalies {
            background: #1a1a2e;
            border: 1px solid #ff6b6b;
            border-radius: 8px;
            padding: 20px;
            margin-top: 30px;
        }
        .anomaly-title {
            color: #ff6b6b;
            font-size: 1.5em;
            margin-bottom: 15px;
        }
        .anomaly-item {
            background: #16213e;
            border-left: 4px solid #ff6b6b;
            padding: 15px;
            margin-bottom: 10px;
            border-radius: 4px;
        }
        .anomaly-type {
            color: #ff6b6b;
            font-weight: bold;
            text-transform: uppercase;
            font-size: 0.9em;
        }
        .anomaly-desc {
            color: #cccccc;
            margin-top: 5px;
        }
        .severity-high { border-left-color: #ff0000; }
        .severity-medium { border-left-color: #ffaa00; }
        .severity-low { border-left-color: #ffff00; }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚡ CPU Performance Heatmap</h1>
        <div class="subtitle">Process: {{.ProcessName}} (PID: {{.PID}}) | Duration: {{printf "%.1f" .TotalDuration}}s | Window Size: {{printf "%.1f" .WindowSize}}s</div>
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-label">Total Samples</div>
                <div class="stat-value">{{.TotalSamples}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Time Windows</div>
                <div class="stat-value">{{len .TimeWindows}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Unique Functions</div>
                <div class="stat-value">{{len .Functions}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Active Threads</div>
                <div class="stat-value">{{len .Threads}}</div>
            </div>
        </div>

        <div class="chart-container">
            <div class="chart-title">Function Activity Heatmap (Top 30 Functions over Time)</div>
            <div id="heatmap"></div>
        </div>

        <div class="chart-container">
            <div class="chart-title">Kernel vs Userland Distribution</div>
            <div id="kernel-userland-chart"></div>
        </div>

        <div class="chart-container">
            <div class="chart-title">Thread Activity Over Time</div>
            <div id="thread-chart"></div>
        </div>

        <div class="chart-container">
            <div class="chart-title">Sample Count per Time Window</div>
            <div id="samples-chart"></div>
        </div>

        {{if .Anomalies}}
        <div class="anomalies">
            <div class="anomaly-title">⚠️ Detected Anomalies</div>
            {{range .Anomalies}}
            <div class="anomaly-item severity-{{.Severity}}">
                <div class="anomaly-type">{{.Type}}</div>
                <div class="anomaly-desc">Window #{{.WindowIndex}}: {{.Description}}</div>
            </div>
            {{end}}
        </div>
        {{end}}
    </div>

    <script>
        const data = {{.DataJSON}};
        const patterns = {{.PatternsJSON}};

        // Prepare heatmap data - top 30 functions
        function prepareHeatmapData() {
            const functionTotals = {};
            data.time_windows.forEach(window => {
                for (const [fn, count] of Object.entries(window.function_counts || {})) {
                    functionTotals[fn] = (functionTotals[fn] || 0) + count;
                }
            });

            const sortedFunctions = Object.entries(functionTotals)
                .sort((a, b) => b[1] - a[1])
                .slice(0, 30)
                .map(([fn]) => fn);

            const zData = sortedFunctions.map(fn => {
                return data.time_windows.map(window => window.function_counts[fn] || 0);
            });

            const xLabels = data.time_windows.map((w, i) => 
                "W" + i + "<br>" + w.start_time.toFixed(1) + "s"
            );

            return {
                z: zData,
                x: xLabels,
                y: sortedFunctions.map(fn => fn.length > 50 ? fn.substring(0, 47) + "..." : fn),
                type: 'heatmap',
                colorscale: [
                    [0, '#0f0f23'],
                    [0.2, '#1a1a2e'],
                    [0.4, '#16213e'],
                    [0.6, '#0f4c75'],
                    [0.8, '#3282b8'],
                    [1, '#00ff00']
                ],
                hovertemplate: 'Function: %{y}<br>Window: %{x}<br>Samples: %{z}<extra></extra>'
            };
        }

        // Plot function heatmap
        Plotly.newPlot('heatmap', [prepareHeatmapData()], {
            paper_bgcolor: '#1a1a2e',
            plot_bgcolor: '#1a1a2e',
            font: { color: '#cccccc' },
            xaxis: { title: 'Time Window', gridcolor: '#2a2a3e' },
            yaxis: { title: 'Function', gridcolor: '#2a2a3e', automargin: true },
            height: 800
        }, {responsive: true});

        // Kernel vs Userland
        const kernelData = data.time_windows.map(w => w.kernel_percent);
        const userlandData = data.time_windows.map(w => w.userland_percent);
        const windowLabels = data.time_windows.map((w, i) => i);

        Plotly.newPlot('kernel-userland-chart', [
            {
                x: windowLabels,
                y: kernelData,
                name: 'Kernel',
                type: 'scatter',
                fill: 'tozeroy',
                line: { color: '#ff6b6b' }
            },
            {
                x: windowLabels,
                y: userlandData,
                name: 'Userland',
                type: 'scatter',
                fill: 'tozeroy',
                line: { color: '#00ff00' }
            }
        ], {
            paper_bgcolor: '#1a1a2e',
            plot_bgcolor: '#1a1a2e',
            font: { color: '#cccccc' },
            xaxis: { title: 'Time Window', gridcolor: '#2a2a3e' },
            yaxis: { title: 'Percentage %', gridcolor: '#2a2a3e' },
            height: 400
        }, {responsive: true});

        // Thread activity
        const threads = data.threads;
        const threadTraces = threads.slice(0, 10).map(tid => {
            return {
                x: windowLabels,
                y: data.time_windows.map(w => w.thread_counts[tid] || 0),
                name: 'TID ' + tid,
                type: 'scatter',
                mode: 'lines'
            };
        });

        Plotly.newPlot('thread-chart', threadTraces, {
            paper_bgcolor: '#1a1a2e',
            plot_bgcolor: '#1a1a2e',
            font: { color: '#cccccc' },
            xaxis: { title: 'Time Window', gridcolor: '#2a2a3e' },
            yaxis: { title: 'Samples', gridcolor: '#2a2a3e' },
            height: 400
        }, {responsive: true});

        // Samples per window
        Plotly.newPlot('samples-chart', [{
            x: windowLabels,
            y: data.time_windows.map(w => w.sample_count),
            type: 'bar',
            marker: { color: '#00ff00' }
        }], {
            paper_bgcolor: '#1a1a2e',
            plot_bgcolor: '#1a1a2e',
            font: { color: '#cccccc' },
            xaxis: { title: 'Time Window', gridcolor: '#2a2a3e' },
            yaxis: { title: 'Sample Count', gridcolor: '#2a2a3e' },
            height: 400
        }, {responsive: true});
    </script>
</body>
</html>`

	tmpl, err := template.New("heatmap").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Prepare data for template
	dataJSON, _ := json.Marshal(data)
	patternsJSON, _ := json.Marshal(patterns)

	templateData := struct {
		*HeatmapData
		Anomalies    []Anomaly
		DataJSON     template.JS
		PatternsJSON template.JS
	}{
		HeatmapData:  data,
		Anomalies:    patterns.Anomalies,
		DataJSON:     template.JS(dataJSON),
		PatternsJSON: template.JS(patternsJSON),
	}

	outputPath := filepath.Join(outputDir, "heatmap.html")
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating HTML file: %v", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, templateData); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	fmt.Printf("✓ Interactive heatmap saved to: %s\n", outputPath)
	return nil
}

