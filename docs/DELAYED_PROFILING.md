# Delayed Profiling Guide

This guide explains how to use delayed profiling features for benchmark integration and iterative testing workflows.

## Overview

Delayed profiling allows you to:
- **Exclude warm-up periods** from performance analysis
- **Profile specific time windows** within long-running processes
- **Integrate with benchmark frameworks** that require warm-up
- **Automate profiling** in testing pipelines

## Basic Usage

### Exclude Warm-up Period

Wait 30 seconds before starting capture:

```bash
sudo blc-perf-analyzer \
  --process myapp \
  --delay-start 30 \
  --duration 60
```

This will:
1. Find the process
2. Wait 30 seconds (process warm-up)
3. Capture for 60 seconds
4. Generate analysis

### Fixed Profiling Window

Use `--profile-window` for clarity when working with delays:

```bash
sudo blc-perf-analyzer \
  --process mariadbd \
  --delay-start 30 \
  --profile-window 60 \
  --generate-flamegraph
```

Timeline:
```
t=0s     : Command starts, finds process
t=0-30s  : Delay period (warm-up)
t=30s    : Profiling starts
t=30-90s : Active profiling
t=90s    : Analysis complete
```

## Automation Integration

### Quiet Mode

For scripting, use `--quiet` to get only the result directory:

```bash
RESULT_DIR=$(sudo blc-perf-analyzer \
  --process postgres \
  --delay-start 10 \
  --profile-window 30 \
  --quiet)

echo "Analysis saved to: $RESULT_DIR"
```

### Custom Output Directory

Specify output location for organized results:

```bash
sudo blc-perf-analyzer \
  --process myapp \
  --delay-start 30 \
  --profile-window 60 \
  --output-dir /var/log/perf/test-run-001 \
  --generate-heatmap \
  --quiet
```

### Exit Codes

The tool uses standard exit codes for automation:

- `0`: Success
- `1`: Configuration error
- `1`: Process not found
- `1`: Permission denied
- `1`: Capture failed

## Benchmark Framework Integration

### Generic Pattern

```bash
#!/bin/bash

# Start your benchmark
./start_benchmark.sh &
BENCH_PID=$!

# Profile after warm-up
sudo blc-perf-analyzer \
  --pid $BENCH_PID \
  --delay-start 30 \
  --profile-window 60 \
  --output-dir ./results/$(date +%Y%m%d-%H%M%S) \
  --generate-flamegraph \
  --generate-heatmap \
  --quiet

# Wait for benchmark to complete
wait $BENCH_PID
```

### With Process Name

```bash
#!/bin/bash

# Start database
systemctl start mariadb

# Run benchmark workload
sysbench --test=oltp prepare &

# Wait for initialization
sleep 10

# Start profiling after warm-up
sudo blc-perf-analyzer \
  --process mariadbd \
  --delay-start 30 \
  --profile-window 120 \
  --output-dir ./results/sysbench-run \
  --generate-flamegraph

# Continue with test...
```

### Multiple Iterations

```bash
#!/bin/bash

for i in {1..5}; do
  echo "Run $i of 5"
  
  # Start workload
  ./benchmark.sh &
  BENCH_PID=$!
  
  # Profile steady state
  RESULT=$(sudo blc-perf-analyzer \
    --pid $BENCH_PID \
    --delay-start 30 \
    --profile-window 60 \
    --output-dir ./results/iteration-$i \
    --generate-heatmap \
    --quiet)
  
  echo "Results: $RESULT"
  
  # Wait for completion
  wait $BENCH_PID
  
  # Cool down
  sleep 10
done
```

## Process Liveness Checking

The tool monitors process health during the delay period:

```bash
sudo blc-perf-analyzer \
  --process short_lived_app \
  --delay-start 60 \
  --duration 30
```

If the process terminates during the 60-second delay, you'll get:
```
Error: process terminated during delay period (after 23 seconds)
```

Progress indicators (non-quiet mode):
```
Found process 'myapp' with PID: 12345
Waiting 30 seconds before starting capture...
  ... 5/30 seconds elapsed
  ... 10/30 seconds elapsed
  ... 15/30 seconds elapsed
  ... 20/30 seconds elapsed
  ... 25/30 seconds elapsed
Starting capture now...
Capturing CPU profile for 60 seconds (PID: 12345)...
```

## Best Practices

### Warm-up Duration

Choose appropriate delay based on your application:

- **Databases**: 30-60 seconds for cache warm-up
- **Web servers**: 10-30 seconds for connection pool initialization
- **Batch processors**: Minimal or none
- **JVM applications**: 60+ seconds for JIT compilation

### Profiling Window

Balance between:
- **Shorter windows** (30-60s): Faster results, may miss patterns
- **Longer windows** (120-300s): Better statistical significance

### Heatmap Window Size

Adjust based on profiling duration:

```bash
# Short capture with fine granularity
--profile-window 30 --heatmap-window-size 0.5

# Long capture with coarser granularity
--profile-window 300 --heatmap-window-size 2.0
```

### Error Handling

```bash
#!/bin/bash

set -e

RESULT=$(sudo blc-perf-analyzer \
  --process myapp \
  --delay-start 30 \
  --profile-window 60 \
  --output-dir ./results \
  --quiet 2>&1)

if [ $? -eq 0 ]; then
  echo "Success: $RESULT"
  # Process results...
else
  echo "Failed: $RESULT"
  exit 1
fi
```

## Example: Database Benchmark

Complete example profiling a database during benchmark:

```bash
#!/bin/bash

DB_PROCESS="mariadbd"
WARM_UP=30
PROFILE_TIME=120
RESULTS_BASE="./perf-results"

# Ensure database is running
if ! pgrep -x "$DB_PROCESS" > /dev/null; then
  echo "Error: $DB_PROCESS not running"
  exit 1
fi

# Create results directory
mkdir -p "$RESULTS_BASE"
RUN_ID=$(date +%Y%m%d-%H%M%S)

# Start benchmark workload
echo "Starting benchmark workload..."
./benchmark.sh &
BENCH_PID=$!

# Profile database after warm-up
echo "Profiling $DB_PROCESS (warm-up: ${WARM_UP}s, capture: ${PROFILE_TIME}s)"
RESULT=$(sudo blc-perf-analyzer \
  --process "$DB_PROCESS" \
  --delay-start "$WARM_UP" \
  --profile-window "$PROFILE_TIME" \
  --output-dir "$RESULTS_BASE/run-$RUN_ID" \
  --generate-flamegraph \
  --generate-heatmap \
  --heatmap-window-size 1.0)

if [ $? -eq 0 ]; then
  echo "Profiling complete: $RESULT"
  
  # Generate summary report
  cat "$RESULT/summary.txt"
  
  # Check for anomalies
  if [ -f "$RESULT/patterns.json" ]; then
    ANOMALIES=$(jq '.anomalies | length' "$RESULT/patterns.json")
    echo "Detected $ANOMALIES performance anomalies"
  fi
else
  echo "Profiling failed"
  kill $BENCH_PID 2>/dev/null
  exit 1
fi

# Wait for benchmark
echo "Waiting for benchmark to complete..."
wait $BENCH_PID

echo "Done. Results in: $RESULT"
```

## Troubleshooting

### Process Dies During Delay

```
Error: process terminated during delay period (after 15 seconds)
```

**Solution**: Reduce `--delay-start` or ensure process is stable.

### Delay Too Long

```
Error: process with PID 12345 no longer exists
```

**Solution**: Verify process runs longer than delay period.

### Permission Issues

```
Error: error running perf: Access denied
```

**Solution**: Run with `sudo` or configure CAP_PERFMON (see main README).

## See Also

- [Main README](../README.md) - General usage
- [Examples](EXAMPLES.md) - More usage examples
- [Roadmap](../ROADMAP.md) - Upcoming features
