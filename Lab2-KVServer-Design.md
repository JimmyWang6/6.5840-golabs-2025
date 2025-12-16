# Lab 2: Key/Value Server - Design Document

## Overview

This document describes the design and implementation of a versioned key/value storage server with optimistic concurrency control. The system provides a simple distributed storage service where clients can store and retrieve key/value pairs, with version numbers used to detect and prevent conflicting concurrent updates.

## Architecture

### System Components

1. **KV Server** (`server.go`)
   - Single-server storage system
   - Maintains key/value pairs with version numbers
   - Handles concurrent client requests with proper synchronization
   - Implements optimistic concurrency control

2. **KV Clerk/Client** (`client.go`)
   - Client library for interacting with the server
   - Handles RPC communication and retries
   - Implements at-most-once semantics for Put operations
   - Manages uncertainty with ErrMaybe responses

3. **RPC Protocol** (`rpc/rpc.go`)
   - Defines Get and Put operations
   - Error codes: OK, ErrNoKey, ErrVersion, ErrMaybe
   - Version-based concurrency control

4. **Lock Service** (`lock/lock.go`)
   - Distributed lock implementation built on top of KV service
   - Uses optimistic concurrency to implement acquire/release semantics
   - Demonstrates practical use of versioned storage

### Core Data Structures

#### Server State
```go
type KVServer struct {
    mu sync.Mutex
    kv map[string]rpc.PutArgs  // Stores value and version for each key
}
```

Each key stores both its value and current version number:
```go
type PutArgs struct {
    Key     string
    Value   string
    Version Tversion  // uint64
}
```

## Design Principles

### 1. Versioned Storage with Optimistic Concurrency Control

Every key has an associated version number that increments with each successful update:

```go
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
    kv.mu.Lock()
    defer kv.mu.Unlock()
    
    if value, ok := kv.kv[args.Key]; ok {
        // Key exists - check version
        if value.Version == args.Version {
            newArgs := *args
            newArgs.Version = args.Version + 1
            kv.kv[args.Key] = newArgs
            reply.Err = rpc.OK
        } else {
            reply.Err = rpc.ErrVersion  // Version mismatch
        }
    } else {
        // Key doesn't exist - only allow if version is 0
        if args.Version == 0 {
            newArgs := *args
            newArgs.Version = 1
            kv.kv[args.Key] = newArgs
            reply.Err = rpc.OK
        } else {
            reply.Err = rpc.ErrNoKey
        }
    }
}
```

**Version Rules:**
- New keys start at version 1 (request must specify version 0)
- Updates must specify current version; server increments on success
- Mismatched versions return ErrVersion (conflict detected)
- Non-existent keys with version > 0 return ErrNoKey

### 2. At-Most-Once Semantics with Uncertainty

The client must handle network failures and provide at-most-once guarantees for Put operations:

```go
func (ck *Clerk) Put(key, value string, version rpc.Tversion) rpc.Err {
    retry := false
    
    for {
        args := rpc.PutArgs{Key: key, Value: value, Version: version}
        reply := rpc.PutReply{}
        success := ck.clnt.Call(ck.server, "KVServer.Put", &args, &reply)
        
        if success {
            if retry && reply.Err == rpc.ErrVersion {
                // Uncertain: first RPC might have succeeded
                return rpc.ErrMaybe
            } else {
                return reply.Err
            }
        } else {
            // RPC failed, will retry
            retry = true
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

**Key Insight:** If the first Put RPC fails (network issue), but subsequent retry gets ErrVersion, the client cannot determine if:
- The first RPC succeeded (value was updated, but response was lost)
- The first RPC never arrived (another client updated the key)

Therefore, the client returns `ErrMaybe` to indicate uncertainty.

### 3. Error Types and Semantics

| Error | Meaning | Client Action |
|-------|---------|---------------|
| **OK** | Operation succeeded | Continue |
| **ErrNoKey** | Key doesn't exist (Get), or Put tried to update non-existent key with version > 0 | Application decides next action |
| **ErrVersion** | Version mismatch - concurrent update detected | Read new version and retry, or return ErrMaybe if uncertain |
| **ErrMaybe** | Client-side only - uncertain if Put succeeded | Application must handle uncertainty |

### 4. Get Operation

Get is simpler since reads are idempotent:

```go
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
    kv.mu.Lock()
    defer kv.mu.Unlock()
    
    if value, ok := kv.kv[args.Key]; ok {
        reply.Value = value.Value
        reply.Version = value.Version
        reply.Err = rpc.OK
    } else {
        reply.Err = rpc.ErrNoKey
    }
}
```

Client retries Get indefinitely on RPC failure:
```go
func (ck *Clerk) Get(key string) (string, rpc.Tversion, rpc.Err) {
    for {
        args := rpc.GetArgs{Key: key}
        reply := rpc.GetReply{}
        success := ck.clnt.Call(ck.server, "KVServer.Get", &args, &reply)
        if !success {
            continue  // Retry on failure
        }
        return reply.Value, reply.Version, reply.Err
    }
}
```

## Distributed Lock Implementation

The lock service demonstrates a practical application of versioned storage:

### Lock State
- **Unlocked:** key value is empty string `""`
- **Locked:** key value contains lock holder's name

### Acquire Algorithm

```go
func (lk *Lock) Acquire() {
    for {
        res, version, err := lk.ck.Get(lk.l)
        if err == rpc.OK {
            if res != "" {
                // Lock held by someone else, try again
                continue
            } else {
                // Lock appears free, try to acquire
                e := lk.ck.Put(lk.l, lk.name, version)
                if e == rpc.OK {
                    // Successfully acquired
                    lk.version = version
                    break
                } else if e == rpc.ErrMaybe {
                    // Uncertain - verify ownership
                    res2, _, err2 := lk.ck.Get(lk.l)
                    if err2 == rpc.OK && res2 == lk.name {
                        lk.version = version
                        break  // We actually acquired it
                    }
                    // Not ours, retry
                }
                // ErrVersion means someone else got it, retry
            }
        }
        // ErrNoKey means need to initialize, retry
    }
}
```

### Release Algorithm

```go
func (lk *Lock) Release() {
    for {
        // Try to clear the lock with our version
        e := lk.ck.Put(lk.l, "", lk.version+1)
        if e == rpc.OK {
            break  // Successfully released
        } else if e == rpc.ErrMaybe {
            // Uncertain if release succeeded
            res, _, err := lk.ck.Get(lk.l)
            if err == rpc.OK && res == "" {
                break  // Lock is released
            }
            // Still locked, retry
        }
        // ErrVersion or ErrNoKey shouldn't happen if lock held correctly
    }
}
```

**Critical Insight:** The lock uses optimistic concurrency to prevent race conditions. Multiple clients might try to acquire simultaneously, but only one will succeed due to version checking.

## Concurrency Control

### Server-Side Synchronization

The server uses a single mutex to protect the key/value map:

```go
kv.mu.Lock()
defer kv.mu.Unlock()
// All operations on kv.kv happen within critical section
```

This ensures:
- **Atomicity:** Get/Put operations are atomic
- **Consistency:** Version checks and updates are indivisible
- **Isolation:** No concurrent modifications to same key

### Client-Side Retry Logic

Clients handle three scenarios:
1. **RPC Success:** Process server response
2. **RPC Failure (first attempt):** Retry indefinitely
3. **RPC Failure (retry):** Track that this is a retry for Put operations

The `retry` flag in Put is crucial for detecting uncertainty.

## Testing Strategy

### Basic Functionality Tests
- Single client Put and Get operations
- Version number correctness
- ErrVersion on concurrent updates
- ErrNoKey for non-existent keys

### Concurrent Access Tests
```go
// Multiple clients racing to update same key
func TestPutConcurrentReliable(t *testing.T) {
    // Spawn NCLNT clients
    // Each tries to Put to same key
    // Verify linearizability with Porcupine
}
```

### Memory Management Tests
```go
// Verify server doesn't leak memory with many clients
func TestMemPutManyClientsReliable(t *testing.T) {
    // Create 100,000 clients
    // Perform operations
    // Check memory usage is reasonable
}
```

### Unreliable Network Tests
- Put operations with dropped RPCs
- Verify ErrMaybe is returned appropriately
- Get operations retry correctly
- Lock service works despite network failures

### Linearizability Verification

The tests use [Porcupine](https://github.com/anishathalye/porcupine) to verify linearizability:
- Records all client operations and responses
- Checks if there exists a valid sequential ordering
- Ensures version numbers form a consistent total order

## Performance Considerations

### 1. Single Lock for Simplicity

The server uses one mutex for all operations:
- **Pros:** Simple, correct, no deadlock risk
- **Cons:** Limits scalability - all operations serialize

**For Lab 2:** This is acceptable since we have only one server.

**Future:** Could use:
- Per-key locking (finer granularity)
- Lock-free data structures
- Read-write locks (optimize for read-heavy workloads)

### 2. Memory Efficiency

The server stores complete PutArgs (key, value, version) for each entry:
```go
kv map[string]rpc.PutArgs
```

**Improvement:** Could separate storage:
```go
type Entry struct {
    Value   string
    Version uint64
}
kv map[string]Entry
```

This avoids redundantly storing the key.

### 3. Client Retry Delays

Clients sleep 100ms between retries:
- Prevents overwhelming server with rapid retries
- Balances latency vs. resource usage

**Tuning:** Could implement exponential backoff for better behavior under sustained failures.

## Design Trade-offs

### Optimistic vs. Pessimistic Concurrency

**Optimistic (Chosen):**
- Clients read version, then attempt update
- Server rejects if version changed
- Good for low-contention workloads

**Pessimistic (Alternative):**
- Clients acquire locks before updates
- Blocks other clients
- Better for high-contention scenarios

**Justification:** Optimistic control is simpler and sufficient for the lab's requirements.

### At-Most-Once vs. Exactly-Once

The system provides **at-most-once** with uncertainty (ErrMaybe):
- Simpler than exactly-once (no duplicate detection)
- Applications must handle ErrMaybe appropriately
- Sufficient for many use cases (lock acquire/release handles it)

**Exactly-once would require:**
- Client request IDs
- Server-side duplicate detection tables
- More complex state management

## Common Pitfalls

1. **Forgetting to increment version:** Server must return version+1 on success
2. **Not tracking retry state:** Client must distinguish first RPC from retries
3. **Ignoring ErrMaybe:** Applications must handle uncertainty correctly
4. **Race conditions in lock:** Must verify lock ownership after ErrMaybe
5. **Holding mutex during RPC:** Server should never make RPCs while holding locks

## Extensions and Future Work

### Possible Enhancements

1. **Append Operation:**
   ```go
   Append(key, value string) error
   ```
   Useful for log-like data structures

2. **Conditional Put:**
   ```go
   PutIf(key, value, oldValue string) error
   ```
   Update only if current value matches expected

3. **Multi-Key Transactions:**
   ```go
   Transaction(ops []Operation) error
   ```
   Atomic updates to multiple keys

4. **Replication:**
   - Use Raft (Lab 3) for fault tolerance
   - Multiple servers for high availability

5. **Persistence:**
   - Write-ahead logging
   - Periodic checkpoints
   - Recovery from crashes

## Connection to Later Labs

This simple key/value server is the foundation for:

- **Lab 3 (Raft):** Replicate the KV server for fault tolerance
- **Lab 4 (Sharded KV):** Partition keys across multiple Raft groups
- **Lab 5 (Transactions):** Add multi-key atomic operations

The versioning and at-most-once semantics learned here are crucial for building correct distributed systems.

## References

- [Optimistic Concurrency Control](https://en.wikipedia.org/wiki/Optimistic_concurrency_control)
- [Linearizability](https://cs.brown.edu/~mph/HerlihyW90/p463-herlihy.pdf)
- [Porcupine Linearizability Checker](https://github.com/anishathalye/porcupine)
- [MIT 6.5840 Lab 2 Specification](https://pdos.csail.mit.edu/6.824/labs/lab-kvsrv.html)

