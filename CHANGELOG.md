# Changelog

All notable changes to BLC Perf Analyzer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Delayed profiling start** (`--delay-start`) for excluding warm-up periods in benchmarks
- **Profile window** (`--profile-window`) as explicit alternative to `--duration`
- **Quiet mode** (`--quiet` / `-q`) for minimal output in automated scenarios
- **Custom output directory** (`--output-dir`) for explicit result path control
- **Process liveness checking** during delay period to detect early termination
- **Enhanced help documentation** with organized flag categories

### Changed
- Improved CLI flag organization (target, timing, output, analysis)
- Better output messages with capture progress indicators
- Refactored capture logic to support delayed start workflow

## [1.0.0] - 2024-12-16

### Added

#### Core Functionality
- **Automated perf capture** with configurable duration
- **Process targeting** by name (`--process`) or PID (`--pid`)
- **System detection** for automatic perf installation
- **Permission checking** for kernel profiling capabilities
- **Organized output** in timestamped directories

#### Analysis Features
- **Flamegraph generation** using Brendan Gregg's scripts
  - Automatic download if not installed
  - SVG output with interactive features
  - Folded stack trace export
- **Interactive temporal heatmaps** (HTML/Plotly.js)
  - Function activity over time visualization
  - Top 30 functions heatmap
  - Kernel vs userland distribution graphs
  - Thread activity timelines
  - Sample count distribution charts
- **Performance classification system**
  - Kernel core functions detection
  - Kernel driver/module identification
  - LibC function classification
  - LibPthread (threading) detection
  - MySQL/MariaDB library identification
  - Application binary categorization
- **Pattern detection engine**
  - Lock contention identification (mutex/futex)
  - High syscall/kernel activity detection
  - CPU spike analysis
  - Anomaly detection with severity levels

#### Parser & Data Processing
- **Full perf script parser** (`internal/parser`)
  - Robust regex-based parsing
  - Stack frame extraction
  - Timestamp and thread tracking
  - CPU core assignment
- **Temporal windowing system**
  - Configurable time window sizes
  - Sample partitioning by time
  - Window-based statistical analysis
- **Statistics generation**
  - Top functions by sample count
  - Category distribution (userland/kernel)
  - Per-thread activity metrics
  - Percentage calculations

#### Output Formats
- **JSON reports**
  - `summary.json`: Overall statistics
  - `heatmap-data.json`: Full temporal data
  - `patterns.json`: Detected anomalies
- **Text reports**
  - `summary.txt`: Human-readable summary
  - `perf-report.txt`: Detailed perf output
  - `perf-output.txt`: Raw perf script data
- **Visual outputs**
  - `flamegraph.svg`: Interactive flamegraph
  - `heatmap.html`: Interactive temporal visualization

#### Testing
- **Parser tests** (`internal/parser/perfscript_test.go`)
  - Sample parsing validation
  - Frame classification tests
  - Time partitioning verification
  - Benchmark tests for performance
- **Heatmap tests** (`internal/heatmap/generator_test.go`)
  - Pattern detection validation
  - HTML generation tests
  - Anomaly detection verification
  - Benchmark tests

#### Documentation
- **Comprehensive README.md**
  - Feature overview
  - Installation instructions
  - Usage examples
  - Output format documentation
  - Architecture description
  - Contributing guidelines
- **CHANGELOG.md** (this file)
- **MIT LICENSE**

#### Command-Line Interface
- `--process, -p`: Target process by name
- `--pid`: Target process by PID
- `--duration, -d`: Capture duration (default: 30s)
- `--generate-flamegraph`: Enable flamegraph generation
- `--generate-heatmap`: Enable heatmap generation
- `--heatmap-window-size`: Configure time window size (default: 1.0s)

### Technical Details

#### Dependencies
- `github.com/spf13/cobra`: CLI framework
- `plotly.js` (CDN): Interactive visualizations
- Go 1.19+: Language runtime

#### System Requirements
- Linux kernel with perf support
- Root/sudo access for profiling
- Sufficient disk space for captures

#### Architecture
```
cmd/blc-perf-analyzer/     CLI entry point
internal/analysis/         Report generation & aggregation
internal/capture/          Perf execution & data capture
internal/detector/         System detection & setup
internal/heatmap/          Temporal heatmap generation
internal/parser/           Perf script parsing
internal/process/          Process utilities
```

### Security
- Validates all user input
- Checks permissions before execution
- No hardcoded paths or credentials
- Safe file operations with error handling

### Performance
- Efficient regex-based parsing
- Minimal memory footprint during capture
- Parallel-ready architecture
- Benchmarked critical paths

### Known Limitations
- Linux-only (no macOS/Windows)
- Requires root/sudo
- Single-process captures only
- Large files (>1GB) may be slow

---

## Future Versions (Roadmap)

### [1.1.0] - Planned
- Web UI mode with live updates
- Comparative analysis (diff two captures)
- Export to Prometheus/InfluxDB
- Multi-process capture support

### [1.2.0] - Planned
- Off-CPU analysis
- Memory profiling integration
- Container/Docker awareness
- Database-specific analysis modes

### [2.0.0] - Future
- Distributed tracing integration
- Machine learning anomaly detection
- Historical trend analysis
- Cloud deployment support

---

## Release Notes

### Version 1.0.0 - Initial Release

This is the first stable release of BLC Perf Analyzer. It provides enterprise-grade CPU performance analysis for Linux systems with a focus on:

1. **Ease of Use**: Single command captures and analyzes
2. **Rich Visualizations**: Flamegraphs and temporal heatmaps
3. **Intelligent Analysis**: Automatic classification and pattern detection
4. **Professional Quality**: Comprehensive tests, documentation, and error handling

The tool has been designed for production use by SREs, DBAs, and performance engineers who need reliable, repeatable performance profiling.

**Tested on:**
- Ubuntu 20.04, 22.04
- Debian 11, 12
- RHEL 8, 9
- CentOS 8
- Alpine Linux 3.18+

**Typical Use Cases:**
- Database performance tuning (MariaDB, PostgreSQL, MySQL)
- Application server profiling (NGINX, Apache)
- System service analysis
- Production incident investigation
- Capacity planning
- Performance regression testing

---

[Unreleased]: https://github.com/santiagolertora/blc-perf-analyzer/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/santiagolertora/blc-perf-analyzer/releases/tag/v1.0.0

