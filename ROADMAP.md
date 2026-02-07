# BLC Perf Analyzer - Roadmap

This document outlines the planned features and improvements for BLC Perf Analyzer.

## Phase 1 - Core Usability (Current Focus)

### 1. Delayed Profiling & Fixed Time Windows
**Status:** In Progress  
**Priority:** High

Enable profiling of specific time windows within long-running processes, excluding warm-up periods and focusing on steady-state performance.

**Features:**
- `--delay-start` flag to wait before starting capture
- `--profile-window` as explicit alternative to `--duration`
- Process liveness validation during delay period
- Non-interactive mode with countdown indicator

**Use Cases:**
- Benchmark frameworks requiring warm-up exclusion
- Iterative testing scenarios
- Performance regression testing
- Load testing analysis

### 2. Non-root Execution
**Status:** Planned  
**Priority:** High

Support profiling without root privileges using Linux capabilities.

**Features:**
- CAP_PERFMON capability detection and usage
- Doctor mode for system validation
- Automated remediation suggestions
- Documentation for capability setup

### 3. Symbol Resolution Improvements
**Status:** Planned  
**Priority:** Medium

Enhance symbol resolution to reduce unknown functions in stripped binaries.

**Features:**
- Automatic debug symbol detection
- Installation guides per distribution
- Symbol server support
- Fallback to addr2line when needed

### 4. Compare Mode
**Status:** Planned  
**Priority:** High

Compare two profiling runs to identify performance differences.

**Features:**
- `compare <run-a> <run-b>` command
- Delta analysis for top functions
- Syscall and kernel/userland distribution comparison
- JSON diff output for automation

### 5. Non-interactive Mode
**Status:** Planned  
**Priority:** Medium

Enable silent operation for CI/CD integration.

**Features:**
- `--quiet` flag for minimal output
- Structured exit codes
- JSON-only output option
- `--output-dir` for explicit path control

## Phase 2 - Intelligence Layer

### 6. Advanced Triggers
**Status:** Planned  
**Priority:** Medium

Extend delayed profiling with intelligent start conditions.

**Features:**
- CPU threshold-based triggering
- External signal support (file, HTTP endpoint)
- Framework integration hooks
- Conditional capture logic

### 7. Pattern Detection Engine
**Status:** Partially Complete  
**Priority:** Medium

Current pattern detection includes lock contention, syscall storms, and CPU spikes. Planned enhancements:

**Features:**
- Spinlock detection
- Memory allocator pressure analysis
- Database vs application layer classification
- Configurable detection thresholds

### 8. Database Engine Presets
**Status:** Planned  
**Priority:** Medium

Predefined profiling configurations optimized for specific database engines.

**Features:**
- `--preset mariadb|postgresql|mysql` flag
- Engine-specific symbol classification
- Custom pattern detection rules
- Optimized capture parameters

## Phase 3 - Container & Remote Support

### 9. Container Awareness
**Status:** Planned  
**Priority:** Medium

Native support for containerized workloads.

**Features:**
- Docker container detection
- Namespace-aware profiling
- cgroup correlation
- Container-specific symbol resolution

### 10. Remote Profiling
**Status:** Planned  
**Priority:** Low

Capture profiles from remote systems.

**Features:**
- Agent-based collector
- SSH-based capture
- Centralized report aggregation
- Multi-host correlation

## Phase 4 - Extended Database Support

### 11. ScyllaDB Support
**Status:** Planned  
**Priority:** Low

Specialized analysis for ScyllaDB workloads.

**Features:**
- Reactor stall detection
- Compaction dominance analysis
- Shard imbalance identification
- jemalloc pressure tracking

### 12. Cassandra Support
**Status:** Planned  
**Priority:** Low

JVM and native performance correlation.

**Features:**
- JVM vs native CPU breakdown
- GC correlation analysis
- Compaction thread tracking
- tpstats alignment

### 13. MongoDB Support
**Status:** Planned  
**Priority:** Low

WiredTiger-specific profiling.

**Features:**
- Eviction analysis
- BSON serialization tracking
- Index build impact measurement
- Engine vs query layer separation

## Phase 5 - Integrations & Observability

### 14. CI/CD Integration
**Status:** Planned  
**Priority:** Medium

Examples and tooling for continuous integration.

**Features:**
- GitHub Actions workflow examples
- GitLab CI templates
- Performance regression detection
- Automated report generation

### 15. Monitoring Integration
**Status:** Planned  
**Priority:** Low

Export metrics to observability platforms.

**Features:**
- Prometheus exporter
- Grafana dashboard templates
- Alert threshold configuration
- Time-series data export

## Phase 6 - Advanced Features

### 16. System-level Correlation
**Status:** Planned  
**Priority:** Low

Correlate CPU profiling with other system metrics.

**Features:**
- I/O correlation analysis
- NUMA awareness
- Detailed futex analysis
- Memory pressure correlation

### 17. Security & Compliance
**Status:** Planned  
**Priority:** Low

Security scanning and compliance features.

**Features:**
- SBOM generation
- Dependency vulnerability scanning
- Secure artifact signing
- Audit logging

---

## Contributing

Feature requests and contributions are welcome. Please open an issue to discuss major changes before implementing them.

## Release Planning

- **v1.1.0**: Phase 1 features (delayed profiling, compare mode, non-root)
- **v1.2.0**: Phase 2 features (advanced triggers, presets)
- **v1.3.0**: Phase 3 features (container support, remote profiling)
- **v2.0.0**: Extended database support and integrations

---

**Last Updated:** February 7, 2026  
**Current Version:** v1.0.1
