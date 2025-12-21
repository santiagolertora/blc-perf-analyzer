# BLC Perf Analyzer

<div align="center">

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.19-00ADD8.svg)
![Platform](https://img.shields.io/badge/platform-Linux-lightgrey.svg)

**Enterprise-grade CPU performance analysis tool for Linux systems**

Automated capture and analysis of CPU traces using `perf`, with interactive visualizations and pattern detection.

[Features](#-features) • [Installation](#-installation) • [Usage](#-usage) • [Examples](#-examples) • [Documentation](#-documentation)

</div>

---

## 🎯 Overview

BLC Perf Analyzer is an open-source tool written in Go that automates the complex workflow of CPU performance profiling on Linux systems. It wraps the powerful `perf` tool and provides intelligent analysis, classification, and visualization of performance data.

### When to Use It?

- 🔥 **Production Issues**: Troubleshooting high CPU usage in production or staging environments
- 🗄️ **Database Tuning**: Performance analysis of databases (MariaDB, PostgreSQL, MySQL)
- 🔍 **Bottleneck Detection**: Quickly identifying userland vs. kernel bottlenecks
- 📊 **Performance Reports**: Generating flamegraphs and heatmaps for visualization
- ⚡ **Real-time Analysis**: Understanding process behavior under load without manual perf scripting

### Target Users

- Site Reliability Engineers (SREs)
- Database Administrators (DBAs)
- Performance Engineers
- DevOps Teams
- System Administrators
- Anyone needing deep insights into Linux process performance

---

## ✨ Features

### Core Capabilities

- ✅ **Automatic System Detection**: Detects OS distribution and installs `perf` if needed
- 🔒 **Permission Management**: Verifies and guides on required kernel permissions
- 🎯 **Flexible Targeting**: Analyze by process name or PID
- ⏱️ **Configurable Duration**: Capture from seconds to hours
- 📁 **Organized Output**: Timestamped directories with all analysis artifacts

### Advanced Analysis

- 🔥 **Flamegraph Generation**: Interactive SVG flamegraphs using Brendan Gregg's scripts
- 🌡️ **Temporal Heatmaps**: See CPU usage patterns evolve over time
- 🧠 **Automatic Classification**: Categorizes functions as kernel/userland/libc/pthread/mysql
- 📈 **Pattern Detection**: Identifies lock contention, syscall storms, and CPU spikes
- 🎨 **Interactive Visualizations**: HTML-based heatmaps with Plotly.js
- 📊 **Statistical Analysis**: Top functions, time distribution, thread activity

### Output Formats

- **JSON**: Machine-readable data for integration with other tools
- **Text**: Human-readable summaries and reports
- **SVG**: Interactive flamegraphs
- **HTML**: Interactive temporal heatmaps with multiple views

---

## 🚀 Installation

### Prerequisites

- **OS**: Linux (any distribution)
- **Go**: 1.19 or higher
- **Root/Sudo**: Required for perf capture

### Build from Source

```bash
# Clone the repository
git clone https://github.com/santiagolertora/blc-perf-analyzer.git
cd blcperfanalyzer

# Build
go build -o blc-perf-analyzer cmd/blc-perf-analyzer/main.go

# Optional: Install to system
sudo mv blc-perf-analyzer /usr/local/bin/
```

### Quick Install

```bash
go install github.com/santiagolertora/blc-perf-analyzer/cmd/blc-perf-analyzer@latest
```

---

## 💻 Usage

### Basic Syntax

```bash
blc-perf-analyzer --process <name> [flags]
# or
blc-perf-analyzer --pid <number> [flags]
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--process` | `-p` | string | - | Process name to analyze (e.g., 'mariadbd') |
| `--pid` | - | int | - | Process ID to analyze |
| `--duration` | `-d` | int | 30 | Capture duration in seconds |
| `--generate-flamegraph` | - | bool | false | Generate flamegraph visualization |
| `--generate-heatmap` | - | bool | false | Generate temporal heatmap |
| `--heatmap-window-size` | - | float | 1.0 | Time window size for heatmap (seconds) |

---

## 📚 Examples

### Example 1: Quick CPU Profile

Analyze a MariaDB process for 30 seconds (basic output):

```bash
sudo blc-perf-analyzer --process mariadbd
```

**Output:**
```
blc-perf-analyzer-20231216-143022/
├── perf.data
└── perf-output.txt
```

### Example 2: Full Analysis with Flamegraph

Generate a flamegraph for a 60-second capture:

```bash
sudo blc-perf-analyzer --process nginx --duration 60 --generate-flamegraph
```

**Output:**
```
blc-perf-analyzer-20231216-143022/
├── perf.data
├── perf-report.txt
├── flamegraph.svg      ← Open in browser
├── perf.folded
├── summary.json
└── summary.txt
```

### Example 3: Temporal Heatmap Analysis

Analyze CPU usage patterns over time with 1-second windows:

```bash
sudo blc-perf-analyzer --pid 1234 --duration 120 --generate-heatmap
```

**Output:**
```
blc-perf-analyzer-20231216-143022/
├── perf.data
├── perf-report.txt
├── heatmap.html        ← Open in browser
├── heatmap-data.json
├── patterns.json       ← Detected anomalies
├── summary.json
└── summary.txt
```

### Example 4: Complete Analysis

Get everything - flamegraph, heatmap, and all reports:

```bash
sudo blc-perf-analyzer \
  --process postgres \
  --duration 300 \
  --generate-flamegraph \
  --generate-heatmap \
  --heatmap-window-size 2.0
```

### Example 5: High-Resolution Heatmap

For fast-changing workloads, use smaller time windows:

```bash
sudo blc-perf-analyzer \
  --process redis-server \
  --duration 60 \
  --generate-heatmap \
  --heatmap-window-size 0.5  # 500ms windows
```

---

## 📊 Understanding the Output

### Summary Text (`summary.txt`)

```
Performance Analysis Summary
==========================

Process: mariadbd (PID: 12345)
Duration: 60 seconds
Total Samples: 45234

Time Distribution:
- Userland: 65.3%
- Kernel: 32.1%
- Unknown: 2.6%

Top Functions:
1. pthread_mutex_lock (15.2%)
2. _int_malloc (8.7%)
3. do_syscall_64 (7.3%)
...
```

### Patterns JSON (`patterns.json`)

Automatically detects:
- **Lock Contention**: High mutex/futex activity
- **Syscall Storms**: Excessive kernel time
- **CPU Spikes**: Sudden increases in activity
- **Anomalies**: Unusual patterns with severity levels

```json
{
  "lock_contention_windows": [12, 15, 18],
  "high_syscall_windows": [8, 22],
  "cpu_spikes": [25],
  "anomalies": [
    {
      "window_index": 12,
      "type": "lock_contention",
      "description": "High lock contention detected: 67% of samples",
      "severity": "high",
      "value": 67.3
    }
  ]
}
```

### Interactive Heatmap

The HTML heatmap includes:
- **Function Activity**: Top 30 functions over time
- **Kernel vs Userland**: Distribution timeline
- **Thread Activity**: Per-thread CPU usage
- **Sample Distribution**: Activity intensity per window
- **Anomaly Highlights**: Visual indicators of detected issues

---

## 🔧 Advanced Configuration

### Adjusting Kernel Permissions

If you encounter permission errors:

```bash
# Check current settings
cat /proc/sys/kernel/perf_event_paranoid
cat /proc/sys/kernel/kptr_restrict

# Temporarily allow profiling (requires root)
sudo sysctl -w kernel.perf_event_paranoid=-1
sudo sysctl -w kernel.kptr_restrict=0

# Permanently (add to /etc/sysctl.conf)
kernel.perf_event_paranoid=-1
kernel.kptr_restrict=0
```

### Custom Perf Events

For advanced users who want to modify perf parameters, edit:
```
internal/capture/capture.go
```

---

## 🏗️ Architecture

```
blc-perf-analyzer/
├── cmd/blc-perf-analyzer/     # Main entry point
│   └── main.go
├── internal/
│   ├── analysis/              # Report generation
│   │   └── analyzer.go
│   ├── capture/               # Perf execution
│   │   └── capture.go
│   ├── detector/              # System detection
│   │   └── detector.go
│   ├── heatmap/               # Heatmap generation
│   │   ├── generator.go
│   │   └── generator_test.go
│   ├── parser/                # Perf script parser
│   │   ├── perfscript.go
│   │   └── perfscript_test.go
│   └── process/               # Process utilities
│       └── process.go
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
└── LICENSE
```

---

## 🧪 Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./internal/parser
go test -bench=. ./internal/heatmap
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Commit with clear messages
6. Push to your fork
7. Open a Pull Request

### Code Standards

- Follow Go conventions and idioms
- Add tests for all new features
- Update documentation
- No hardcoded values - use configuration
- Enterprise-grade quality only

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👤 Author

**Santiago Lertora**

- GitHub: [@santiagolertora](https://github.com/santiagolertora)
- Email: santiago@lertora.com

---

## 🙏 Acknowledgments

- [Brendan Gregg](http://www.brendangregg.com/) for FlameGraph scripts and performance methodology
- The Linux `perf` development team
- The Go community

---

## 📖 Additional Resources

- [Linux perf Wiki](https://perf.wiki.kernel.org/)
- [Brendan Gregg's Blog](http://www.brendangregg.com/blog/)
- [FlameGraph Repository](https://github.com/brendangregg/FlameGraph)

---

## 🐛 Known Issues & Roadmap

### Known Issues

- Requires root/sudo for capture
- Linux-only (no macOS/Windows support)
- Large captures (>1GB) may be slow to parse

### Roadmap

- [ ] Web UI mode (`--web`) for live viewing
- [ ] Comparative analysis between two captures
- [ ] Docker/container-aware filtering
- [ ] Prometheus/InfluxDB integration
- [ ] Database-specific profiles (MariaDB/PostgreSQL/MongoDB)
- [ ] Off-CPU analysis support
- [ ] Multi-process capture

---

## 👤 Author

**Santiago Lertora**

- 🌐 Website: [santiagolertora.com](https://www.santiagolertora.com)
- 📧 Email: [santiagolertora@gmail.com](mailto:santiagolertora@gmail.com)
- 💼 GitHub: [@santiagolertora](https://github.com/santiagolertora)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**⭐ Star this repo if you find it useful!**

Made with ❤️ for the SRE and DevOps community

[Report Bug](https://github.com/santiagolertora/blc-perf-analyzer/issues) • [Request Feature](https://github.com/santiagolertora/blc-perf-analyzer/issues)

</div>
