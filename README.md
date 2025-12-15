# MIT 6.5840 (6.824) Distributed Systems Labs 2025

This repository contains my implementation of the distributed systems labs from MIT's 6.5840 (formerly 6.824) course.

## Course Information

**Course:** MIT 6.5840 - Distributed Systems  
**Semester:** 2025  
**Language:** Go

## Labs Overview

### 📝 [Lab 1: MapReduce](./Lab1-Design.md)

A distributed MapReduce implementation featuring:
- **Coordinator-Worker Architecture**: Master-worker pattern for task orchestration
- **Fault Tolerance**: Heartbeat-based worker failure detection and task reassignment
- **Parallel Processing**: Concurrent execution of map and reduce tasks
- **RPC Communication**: Unix domain socket-based inter-process communication

**Key Features:**
- Automatic task distribution and load balancing
- Worker crash recovery with task re-execution
- Intermediate file management with partitioning
- Phase-based execution (Map → Reduce → Finish)

**Status:** ✅ Completed - Passes all official test cases

---

### 🔄 Lab 2: Raft Consensus

*Coming soon...*

Implementation of the Raft consensus algorithm for distributed systems.

---

### 🗄️ Lab 3: Key-Value Service

*Coming soon...*

Fault-tolerant key-value storage service built on top of Raft.

---

### 🔀 Lab 4: Sharded Key-Value Service

*Coming soon..*

Distributed sharded key-value service with dynamic re-sharding.

---

## Project Structure

```
├── src/
│   ├── mr/                  # Lab 1: MapReduce implementation
│   │   ├── coordinator.go   # Coordinator (master) logic
│   │   ├── worker.go        # Worker implementation
│   │   └── rpc.go          # RPC definitions
│   ├── raft1/              # Lab 2: Raft consensus
│   ├── kvraft1/            # Lab 3: Key-value service
│   ├── shardkv1/           # Lab 4: Sharded KV service
│   ├── labrpc/             # RPC testing framework
│   ├── labgob/             # Go encoding utilities
│   └── main/               # Entry points and test scripts
├── Lab1-Design.md          # Lab 1 design document
└── README.md               # This file
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Linux/Unix environment (for Unix domain sockets)
- Make

### Building the Project

```bash
cd src/main

# Build MapReduce components
go build -buildmode=plugin ../mrapps/wc.go
go build mrcoordinator.go
go build mrworker.go
go build mrsequential.go
```

### Running MapReduce

#### Sequential Mode (for testing)
```bash
cd src/main
go run mrsequential.go wc.so pg-*.txt
```

#### Distributed Mode

Terminal 1 - Start Coordinator:
```bash
cd src/main
rm mr-out-*
go run mrcoordinator.go pg-*.txt
```

Terminal 2+ - Start Workers:
```bash
cd src/main
go run mrworker.go wc.so
```

### Running Tests

```bash
cd src/mr
go test -run TestBasicMapReduce

# Run all tests
cd src/main
bash test-mr.sh
```

## Implementation Highlights

### Lab 1: MapReduce

- ✅ **Parallel Map Tasks**: Multiple workers process input files concurrently
- ✅ **Parallel Reduce Tasks**: Reduce operations execute in parallel across partitions
- ✅ **Fault Tolerance**: Automatic detection and recovery from worker failures
- ✅ **Task Scheduling**: Efficient pull-based task assignment
- ✅ **Data Partitioning**: Hash-based key distribution for load balancing
- ✅ **Intermediate Storage**: JSON-encoded intermediate files with proper naming

### Test Results

All MapReduce tests passing:
- ✅ Basic MapReduce functionality
- ✅ Word count test
- ✅ Indexer test  
- ✅ Map parallelism
- ✅ Reduce parallelism
- ✅ Job counting
- ✅ Early exit handling
- ✅ Crash recovery

## Design Documents

Detailed design documentation for each lab:

1. **[Lab 1: MapReduce Design](./Lab1-Design.md)** - Complete architecture, data structures, workflow, and fault tolerance mechanisms

## Learning Objectives

Through these labs, I've gained hands-on experience with:

- **Distributed Computing Fundamentals**: Task distribution, parallel processing, fault tolerance
- **RPC Systems**: Remote procedure calls, serialization, network communication
- **Consensus Algorithms**: Raft protocol for distributed agreement
- **Replicated State Machines**: Building fault-tolerant services
- **Distributed Transactions**: Sharding, replication, consistency models
- **Go Concurrency**: Goroutines, channels, mutexes, and concurrent programming patterns

## Resources

- [MIT 6.5840 Course Website](http://nil.csail.mit.edu/6.5840/2024/)
- [Original MapReduce Paper (Google, 2004)](https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf)
- [Raft Consensus Paper](https://raft.github.io/raft.pdf)
- [Go Documentation](https://golang.org/doc/)

## Development Notes

- Code follows Go best practices and idioms
- All critical sections protected with appropriate synchronization
- Comprehensive logging for debugging and monitoring
- Modular design for maintainability and testing

## License

This is coursework for educational purposes. Please refer to MIT's academic integrity policies.

## Acknowledgments

- MIT PDOS (Parallel and Distributed Operating Systems) group
- Course instructors and TAs
- Original MapReduce and Raft paper authors

---

**Note:** This repository contains my personal solutions to the MIT 6.5840 labs. If you're currently taking this course, please follow your institution's academic integrity policies and use this only as a reference after completing your own implementation.

