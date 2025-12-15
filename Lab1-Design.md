# Lab 1: MapReduce - Design Document

## Overview

This document describes the design and implementation of a distributed MapReduce system based on the MIT 6.5840 (formerly 6.824) course. The system consists of a coordinator (master) that orchestrates the work and multiple workers that execute map and reduce tasks in parallel.

## Architecture

The MapReduce system follows a master-worker architecture:

- **Coordinator**: Manages task distribution, tracks worker health, and coordinates the overall job execution
- **Workers**: Execute map and reduce tasks assigned by the coordinator
- **RPC Communication**: Workers communicate with the coordinator through RPC calls for registration, task requests, heartbeats, and completion reports

## Core Data Structures

### 1. Coordinator Structure

```go
type Coordinator struct {
    mu      sync.Mutex        // Protects shared data structures
    State   string            // Current phase: "map", "reduce", or "finish"
    Counter int               // Worker ID counter for registration
    Files   []File            // All map and reduce tasks
    NReduce int               // Number of reduce tasks
    Workers map[int]*WorkerItem // Active workers registry
}
```

**Key Fields:**
- `State`: Tracks the current execution phase (map → reduce → finish)
- `Counter`: Auto-incrementing ID generator for worker registration
- `Files`: Unified task queue containing both map tasks (input files) and reduce tasks
- `Workers`: Registry of all active workers with their metadata

### 2. File (Task) Structure

```go
type File struct {
    Name     string  // File name for map tasks; reduce task ID for reduce tasks
    Index    int     // Unique task index
    Assigned bool    // Whether task is currently assigned
    Done     bool    // Whether task is completed
    Owner    int     // Worker ID that owns this task
    Type     string  // Task type: "map" or "reduce"
}
```

**Purpose:** Represents both map tasks (input files) and reduce tasks in a unified structure.

### 3. Worker Item Structure

```go
type WorkerItem struct {
    WorkerID int       // Unique worker identifier
    NReduce  int       // Number of reduce tasks (for partitioning)
    NMap     int       // Number of map tasks (for reading intermediates)
    LastSeen time.Time // Last heartbeat timestamp
    File     File      // Currently assigned task
}
```

**Purpose:** Tracks worker state and metadata both in coordinator and worker process.

### 4. RPC Message Structures

#### Registration
```go
type RegisterArgs struct {}  // No arguments needed

type RegisterReply struct {
    WorkerID int  // Assigned worker ID
    NReduce  int  // Number of reduce tasks
    NMap     int  // Number of map tasks
}
```

#### Task Reporting
```go
type ReportArgs struct {
    WorkerID int  // Worker identifier
    Type     int  // Reserved for future use
    File     int  // Task index (-1 for initial request)
}

type ReportReply struct {
    Response string  // "doing", "wait", or "done"
    File     File    // Assigned task (if any)
}
```

#### Heartbeat
```go
type HeartbeatArgs struct {
    WorkerID int  // Worker identifier
}

type HeartbeatReply struct {
    Response string  // "ok"
}
```

## System Workflow

### Phase 1: Initialization

1. **Coordinator Startup:**
   - Creates task list: N map tasks (one per input file) + M reduce tasks
   - Initializes all tasks as unassigned and incomplete
   - Starts RPC server on Unix domain socket
   - Launches worker health monitoring goroutine

2. **Worker Registration:**
   - Worker calls `Register()` RPC
   - Coordinator assigns unique worker ID
   - Worker receives `NReduce` and `NMap` parameters
   - Worker starts heartbeat goroutine (every 3 seconds)

### Phase 2: Map Phase

1. **Task Assignment:**
   - Worker requests task via `Report()` with `File=-1` (initial) or completed task index
   - Coordinator finds first unassigned map task
   - Marks task as assigned, sets owner to worker ID
   - Returns task to worker

2. **Map Execution:**
   - Worker reads input file content
   - Applies user-defined `mapf()` function → generates key-value pairs
   - Partitions output into `NReduce` buckets using `ihash(key) % NReduce`
   - Writes intermediate files: `mr-{mapIndex}-{reduceIndex}`
   - Uses JSON encoding for intermediate data

3. **Completion Reporting:**
   - Worker calls `Report()` with completed task index
   - Coordinator marks task as done
   - Coordinator assigns next available map task or returns "wait"

4. **Phase Transition:**
   - When all map tasks are done, coordinator transitions to reduce phase
   - State changes from "map" to "reduce"

### Phase 3: Reduce Phase

1. **Task Assignment:**
   - Coordinator assigns reduce tasks (identified by reduce index)
   - Each reduce task processes one partition across all map outputs

2. **Reduce Execution:**
   - Worker reads all intermediate files: `mr-*-{reduceIndex}`
   - Decodes JSON key-value pairs
   - Sorts all pairs by key
   - Groups values by key
   - Applies user-defined `reducef()` function to each key group
   - Writes final output: `mr-out-{reduceIndex}`

3. **Completion:**
   - Worker reports reduce task completion
   - Coordinator marks reduce task as done

4. **Job Completion:**
   - When all reduce tasks complete, state changes to "finish"
   - Workers receive "done" response and exit
   - Coordinator's `Done()` returns true, system shuts down

## Fault Tolerance

### Worker Failure Detection

**Heartbeat Mechanism:**
- Workers send heartbeat every 3 seconds
- Coordinator tracks `LastSeen` timestamp for each worker
- Scanner goroutine runs every 10 seconds

**Failure Handling:**
- If `now - LastSeen > 10 seconds`: worker declared dead
- Coordinator removes worker from registry
- All assigned but incomplete tasks are reset:
  - `Assigned = false`
  - `Owner = -1`
- Tasks become available for reassignment to other workers

### Task Re-execution

- Failed tasks are automatically reassigned to healthy workers
- MapReduce semantics ensure idempotent operations
- Multiple executions produce same intermediate/output files (overwrite)

## Synchronization and Concurrency

### Critical Sections

All coordinator state modifications are protected by `Coordinator.mu`:
- Task assignment
- Task completion marking
- Worker registration/deletion
- State transitions

### Concurrent Workers

- Multiple workers can execute map tasks in parallel
- Multiple workers can execute reduce tasks in parallel
- No limit on number of concurrent workers
- Intermediate files partitioned by reduce index prevent conflicts

### No Worker-Worker Communication

- Workers only communicate with coordinator
- All data exchange through file system (intermediate files)
- No need for worker-level synchronization

## File Management

### Intermediate Files

**Naming Convention:** `mr-{mapTaskIndex}-{reduceTaskIndex}`

**Example:** 
- 8 input files, 10 reduce tasks → 80 intermediate files
- `mr-3-5`: output from map task 3 for reduce partition 5

### Output Files

**Naming Convention:** `mr-out-{reduceTaskIndex}`

**Example:**
- 10 reduce tasks → 10 output files
- `mr-out-0`, `mr-out-1`, ..., `mr-out-9`

### File Operations

- **Map phase:** Each map task creates `NReduce` intermediate files
- **Reduce phase:** Each reduce task reads `NMap` intermediate files
- All file operations are local (no distributed file system in this implementation)

## State Machine

```
[START]
   ↓
[MAP Phase]
   ├─ Assign map tasks
   ├─ Wait for completion
   └─ All map tasks done?
         ↓ YES
[REDUCE Phase]
   ├─ Assign reduce tasks
   ├─ Wait for completion
   └─ All reduce tasks done?
         ↓ YES
[FINISH]
   ↓
[SHUTDOWN]
```

## Key Design Decisions

### 1. Unified Task Structure

Both map and reduce tasks use the `File` struct with a `Type` field. This simplifies task management but requires type checking during assignment.

### 2. Pull-Based Task Assignment

Workers pull tasks by calling `Report()` rather than coordinator pushing tasks. This simplifies worker failure handling and load balancing.

### 3. File-Based Data Exchange

Intermediate data is stored in files rather than in-memory. This provides fault tolerance and simplifies the implementation at the cost of I/O overhead.

### 4. JSON Serialization

Intermediate key-value pairs use JSON encoding for simplicity, though more efficient serialization formats could be used.

### 5. Index-Based Task Identification

Tasks are identified by index rather than name, allowing unified handling of map (file-based) and reduce (partition-based) tasks.

## Testing

The system passes the official MIT 6.5840 test suite:

1. **Word Count Test:** Basic map-reduce functionality
2. **Indexer Test:** Different reduce function
3. **Map Parallelism Test:** Verifies concurrent map execution
4. **Reduce Parallelism Test:** Verifies concurrent reduce execution
5. **Job Count Test:** Ensures correct task counting
6. **Early Exit Test:** Validates proper completion signaling
7. **Crash Test:** Tests fault tolerance with worker failures

## Performance Characteristics

- **Scalability:** Horizontally scalable with number of workers
- **Parallelism:** Limited by number of map/reduce tasks
- **I/O Bound:** Performance constrained by file system operations
- **Network:** Uses Unix domain sockets (local only)

## Limitations and Future Improvements

### Current Limitations

1. **No Distributed File System:** All nodes must share the same file system
2. **No Speculative Execution:** Slow workers can become stragglers
3. **No Combiner:** No local reduce optimization
4. **Simple Partitioning:** Hash-based partitioning may create imbalanced reduces

### Potential Improvements

1. Implement combiner function for local aggregation
2. Add speculative execution for straggler mitigation
3. Use more efficient serialization (e.g., Protocol Buffers)
4. Implement backup tasks for late-running tasks
5. Add support for distributed file systems
6. Implement locality-aware scheduling
7. Add compression for intermediate files

## Conclusion

This MapReduce implementation demonstrates the core concepts of distributed computing including task distribution, fault tolerance, and parallel processing. While simplified compared to production systems like Hadoop, it provides a solid foundation for understanding distributed data processing frameworks.

