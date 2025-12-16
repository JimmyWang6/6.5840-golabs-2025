# Lab 1: MapReduce - Design Document

## Overview

This document describes the design and implementation of a MapReduce system based on the [Google MapReduce paper](https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf). The implementation provides a distributed framework for processing large datasets across multiple worker machines coordinated by a single coordinator.

## Architecture

### System Components

1. **Coordinator** (`coordinator.go`)
   - Central coordinator that manages task distribution
   - Tracks worker registration and health via heartbeats
   - Maintains task states (unassigned, assigned, completed)
   - Handles task reassignment when workers fail

2. **Worker** (`worker.go`)
   - Executes map and reduce tasks assigned by the coordinator
   - Registers with coordinator on startup
   - Sends periodic heartbeats to maintain liveness
   - Reports task completion back to coordinator

3. **RPC Interface** (`rpc.go`)
   - Defines communication protocol between workers and coordinator
   - Includes Register, Report, and Heartbeat RPCs

### Data Structures

#### Coordinator State
```go
type Coordinator struct {
    mu      sync.Mutex
    State   string              // "map", "reduce", or "finish"
    Counter int                 // Worker ID counter
    Files   []File              // Task queue
    NReduce int                 // Number of reduce tasks
    Workers map[int]*WorkerItem // Active workers
}
```

#### File/Task Representation
```go
type File struct {
    Name     string  // Input filename or task identifier
    Index    int     // Task index
    Assigned bool    // Whether task is assigned
    Done     bool    // Whether task is completed
    Owner    int     // Worker ID that owns this task
    Type     string  // "map" or "reduce"
}
```

## Design Decisions

### 1. Phase-Based Execution

The system operates in three distinct phases:
- **Map Phase**: All map tasks must complete before reduce phase begins
- **Reduce Phase**: Reduce tasks process intermediate data from map outputs
- **Finish Phase**: All work is complete, system can shut down

This ensures correctness by preventing reduce tasks from starting before all intermediate files are ready.

### 2. Worker Registration

Workers register with the coordinator on startup and receive:
- Unique Worker ID
- Number of reduce tasks (NReduce)
- Number of map tasks (NMap)

This allows workers to properly structure intermediate file output and understand the task space.

### 3. Heartbeat Mechanism

Workers send periodic heartbeats to prove liveness:
```go
func heartbeat(w *WorkerItem) {
    for {
        time.Sleep(5 * time.Second)
        // Send heartbeat RPC
    }
}
```

The coordinator tracks `LastSeen` timestamps to detect failed workers and reassign their tasks.

### 4. Task Assignment Strategy

The `assign()` function implements task distribution:
- During map phase: assigns unassigned map tasks
- After all maps complete: transitions to reduce phase
- During reduce phase: assigns reduce tasks
- Returns empty task if no work available (worker should wait)

### 5. Intermediate File Format

Map tasks produce intermediate files following the naming pattern:
```
mr-{mapTaskIndex}-{reduceTaskIndex}
```

Each map task creates `NReduce` intermediate files, one for each reduce partition. Files use JSON encoding for key-value pairs:

```go
// Write to intermediate files
for _, kv := range kva {
    r := ihash(kv.Key) % w.NReduce
    encFiles[r].enc.Encode(&kv)
}
```

### 6. Reduce Task Processing

Reduce tasks:
1. Read all intermediate files for their partition: `mr-*-{reduceIndex}`
2. Sort keys to group values
3. Apply reduce function to each key's values
4. Write final output to `mr-out-{reduceIndex}`

### 7. Fault Tolerance

The system handles worker failures through:
- **Heartbeat monitoring**: Coordinator detects unresponsive workers
- **Task reassignment**: Tasks from failed workers are marked unassigned
- **Idempotent writes**: Map and reduce tasks use atomic file creation

Workers that crash mid-task have their work redone by other workers. The coordinator ensures each task completes exactly once.

### 8. Partitioning Function

Keys are partitioned across reduce tasks using a hash function:
```go
func ihash(key string) int {
    h := fnv.New32a()
    h.Write([]byte(key))
    return int(h.Sum32() & 0x7fffffff)
}
```

This ensures balanced distribution and deterministic routing.

## Implementation Highlights

### Concurrency Control

The coordinator uses a single mutex to protect shared state:
```go
c.mu.Lock()
defer c.mu.Unlock()
```

This simple locking strategy prevents race conditions while maintaining adequate performance for the coordinator's coordination role.

### Worker State Machine

Workers follow a simple request-execute-report loop:
1. Request task from coordinator (Report RPC)
2. Execute map or reduce based on task type
3. Report completion back to coordinator

If no task is available, workers sleep briefly before retrying.

### Map Task Execution

```go
// Read input file
contents := Read(w.File.Name)
// Apply map function
kva := mapf(w.File.Name, contents)
// Partition and write intermediate files
for _, kv := range kva {
    r := ihash(kv.Key) % w.NReduce
    encFiles[r].enc.Encode(&kv)
}
```

### Reduce Task Execution

```go
// Read all intermediate files for this reduce task
// Sort by key
sort.Sort(ByKey(intermediate))
// Apply reduce function to each unique key
for i < len(intermediate) {
    j := i + 1
    for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
        j++
    }
    values := []string{}
    for k := i; k < j; k++ {
        values = append(values, intermediate[k].Value)
    }
    output := reducef(intermediate[i].Key, values)
    // Write to output file
}
```

## Performance Considerations

1. **Parallelism**: Multiple workers can execute tasks concurrently
2. **Load Balancing**: First-come-first-served task assignment naturally balances load
3. **Network Efficiency**: Workers read input files directly rather than streaming through coordinator
4. **Stragglers**: Heartbeat-based failure detection allows quick recovery from slow/failed workers

## Testing

The implementation passes the test suite in `test-mr.sh`:
- Basic word count functionality
- Handling of intermediate file indexing
- Worker crash recovery
- Parallel execution correctness
- Various map/reduce applications (indexer, mtiming, rtiming, etc.)

## Future Enhancements

Potential improvements for production use:
1. **Backup Tasks**: Start duplicate tasks for stragglers near job completion
2. **Locality Optimization**: Prefer assigning tasks to workers near input data
3. **Combiner Functions**: Reduce intermediate data size with local aggregation
4. **Dynamic Worker Scaling**: Add/remove workers based on workload
5. **Better Load Balancing**: Consider worker capacity and historical performance

## References

- [MapReduce: Simplified Data Processing on Large Clusters](https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf)
- [MIT 6.5840 Lab 1 Specification](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)

