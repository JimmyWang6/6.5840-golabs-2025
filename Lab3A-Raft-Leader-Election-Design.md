# Lab 3A: Raft Leader Election 设计文档

## 概述

本实验实现了 Raft 共识算法的第一部分：**Leader 选举（Leader Election）**。Raft 是一个用于管理复制日志的共识算法，它通过选举一个 leader 来协调所有操作，保证分布式系统的一致性。

Lab 3A 专注于实现：
- Raft 服务器的三种角色转换（Follower、Candidate、Leader）
- Leader 选举机制
- 心跳（Heartbeat）机制
- 网络分区下的选举行为

## 核心概念

### Raft 三种角色

1. **Follower（跟随者）**
   - 被动接收来自 Leader 和 Candidate 的 RPC
   - 响应投票请求
   - 如果超时未收到心跳，转换为 Candidate

2. **Candidate（候选人）**
   - 发起选举，请求其他服务器投票
   - 如果获得多数票，成为 Leader
   - 如果收到来自新 Leader 的心跳，转换为 Follower
   - 如果选举超时，开始新一轮选举

3. **Leader（领导者）**
   - 定期发送心跳（空的 AppendEntries RPC）给所有 Follower
   - 如果发现更高的 term，降级为 Follower
   - 负责处理客户端请求（Lab 3B）

### 关键机制

#### Term（任期）
- 每个服务器维护一个 `currentTerm` 计数器
- Term 单调递增
- 服务器间通信时会交换 term：
  - 如果发现更高的 term，立即更新并转换为 Follower
  - 如果收到过期的 term，拒绝请求

#### 选举规则
- **多数派原则**：必须获得**严格多数**（> n/2）的投票才能成为 Leader
- **每个 term 每个服务器只能投一票**
- **先到先得**：在同一 term 中，第一个请求投票的 Candidate 获得投票

#### 选举超时（Election Timeout）
- Follower 等待心跳的超时时间
- 超时后转换为 Candidate 并发起选举
- **随机化**：每个服务器的超时时间在一个范围内随机选择（如 150-300ms）
- 目的：避免多个服务器同时发起选举导致选票分散

#### 心跳机制
- Leader 定期发送空的 AppendEntries RPC 作为心跳
- 心跳间隔应**远小于**选举超时时间（如 50ms vs 150-300ms）
- 目的：
  - 维持 Leader 身份
  - 防止 Follower 超时发起不必要的选举
  - 检测网络分区

## 实现细节

### 数据结构

```go
type Raft struct {
    mu              sync.Mutex          // 保护共享状态的互斥锁
    peers           []*labrpc.ClientEnd // 所有服务器的 RPC 端点
    me              int                 // 本服务器在 peers[] 中的索引
    dead            int32               // 用于 Kill()
    heartbeatCancel context.CancelFunc  // 取消心跳协程
    
    // Raft 状态
    Role            string              // "leader", "candidate", "follower"
    term            int                 // 当前任期
    voteFor         map[int]int         // [term]candidateId - 投票记录
    currentLeader   int                 // 当前 Leader ID
    Logs            []Log               // 日志条目（Lab 3B）
    lastHeartbeat   time.Time           // 上次收到心跳的时间
}
```

### RPC 接口

#### RequestVote RPC
**用途**：Candidate 请求投票

**参数（RequestVoteArgs）**：
- `Term`：Candidate 的当前 term
- `Candidate`：Candidate 的服务器 ID
- `LastLogIndex`：Candidate 最后一条日志的索引（Lab 3B）
- `LastLogTerm`：Candidate 最后一条日志的 term（Lab 3B）

**返回值（RequestVoteReply）**：
- `Term`：接收者的 currentTerm，用于 Candidate 更新自己
- `VoteGranted`：true 表示投票给了 Candidate

**逻辑**：
1. 如果 `args.Term < currentTerm`，拒绝投票
2. 如果 `args.Term > currentTerm`，更新自己的 term，转换为 Follower
3. 如果在当前 term 还没有投票，且 Candidate 的日志至少和自己一样新，投票给它
4. 重置选举超时计时器

#### AppendEntries RPC
**用途**：Leader 发送日志条目（或心跳）

**参数（AppendEntriesArgs）**：
- `Term`：Leader 的当前 term
- `LeaderID`：Leader 的服务器 ID
- `PrevLogIndex`：新日志条目之前的日志索引（Lab 3B）
- `PrevLogTerm`：prevLogIndex 处的日志 term（Lab 3B）
- `Entries`：要存储的日志条目（心跳时为空）
- `LeaderCommit`：Leader 的 commitIndex（Lab 3B）

**返回值（AppendEntriesReply）**：
- `Term`：接收者的 currentTerm
- `Success`：如果 follower 包含匹配 prevLogIndex 和 prevLogTerm 的条目，返回 true

**逻辑**（Lab 3A 简化版）：
1. 如果 `args.Term < currentTerm`，返回 false
2. 如果 `args.Term >= currentTerm`：
   - 更新自己的 term
   - 如果是 Candidate，转换为 Follower
   - 重置选举超时计时器
   - 返回 true

### 核心算法

#### 选举流程（Elect）

```
1. 转换为 Candidate
2. 增加 currentTerm
3. 投票给自己（voteCount = 1）
4. 重置选举超时计时器
5. 并发向所有其他服务器发送 RequestVote RPC

对于每个投票响应：
    如果收到更高的 term：
        更新自己的 term
        转换��� Follower
        退出选举
    
    如果收到投票：
        voteCount++
        如果 voteCount > len(peers)/2：
            转换为 Leader
            开始发送心跳
            退出选举
    
    如果选举超时：
        退出当前选举（将在下次 ticker 触发时重新开始）
```

#### Ticker 主循环

```
while not killed:
    sleep for random time (50-350ms)
    
    检查当前角色：
        如果是 Leader：
            // 心跳由单独的协程处理
            继续
        
        如果是 Candidate：
            发起选举
        
        如果是 Follower：
            如果选举超时（未收到心跳）：
                发起选举
```

#### 心跳发送（sendHeartbeats）

```
while Leader 角色未改变：
    向所有其他服务器发送空的 AppendEntries RPC
    
    对于每个响应：
        如果 reply.Term > currentTerm：
            更新 term
            转换为 Follower
    
    sleep for HeartbeatInterval (如 150ms)
```

### 并发与锁管理

#### 锁的使用原则

1. **读写共享状态必须持有锁**
   - `rf.term`, `rf.Role`, `rf.voteFor`, `rf.lastHeartbeat` 等

2. **RPC Handler 中的锁管理**
   ```go
   func (rf *Raft) RequestVote(args, reply) {
       rf.mu.Lock()
       defer rf.mu.Unlock()
       // ... 处理逻辑
   }
   ```

3. **避免死锁**
   - 不要在持有锁的情况下调用 RPC（RPC 可能阻塞）
   - 发送 RPC 前先复制需要的状态，释放锁
   ```go
   rf.mu.Lock()
   term := rf.term
   me := rf.me
   rf.mu.Unlock()
   
   // 发送 RPC（不持有锁）
   rf.sendRequestVote(server, &RequestVoteArgs{Term: term, Candidate: me}, reply)
   ```

4. **数据竞争检测**
   - 使用 `go test -race` 运行测试
   - 确保所有测试都能通过竞争检测

#### 常见的锁问题

1. **GetState() 未加锁**
   ```go
   // ❌ 错误
   func (rf *Raft) GetState() (int, bool) {
       return rf.term, rf.Role == Leader
   }
   
   // ✅ 正确
   func (rf *Raft) GetState() (int, bool) {
       rf.mu.Lock()
       defer rf.mu.Unlock()
       return rf.term, rf.Role == Leader
   }
   ```

2. **transitTo() 内部解锁**
   ```go
   // ❌ 错误：调用者已持有锁并使用 defer
   func (rf *Raft) transitTo(role string) {
       rf.Role = role
       if role == Leader {
           rf.mu.Unlock()  // 导致双重解锁
       }
   }
   
   // ✅ 正确：由调用者管理锁
   func (rf *Raft) transitTo(role string) {
       rf.Role = role
       if role == Leader {
           // 启动心跳（在 goroutine 中）
           go rf.sendHeartbeats()
       }
   }
   ```

3. **Ticker 中无锁读取**
   ```go
   // ❌ 错误
   if rf.Role == Leader {
       // ...
   }
   
   // ✅ 正确
   rf.mu.Lock()
   role := rf.Role
   rf.mu.Unlock()
   if role == Leader {
       // ...
   }
   ```

### 关键 Bug 与修复

#### Bug 1: 投票判断条件错误

**问题**：
```go
if votes >= len(rf.peers)/2 {  // ❌ 错误
    rf.transitTo(Leader)
}
```

对于 3 个服务器：`votes >= 3/2` 即 `votes >= 1`，意味着 1 票就能成为 leader！

**修复**：
```go
if votes > len(rf.peers)/2 {  // ✅ 正确
    rf.transitTo(Leader)
}
```

3 个服务器需要 > 1.5，即至少 2 票（多数派）。

#### Bug 2: 重复的投票判断逻辑

**问题**：在 `Elect()` 函数中有两处投票判断，导致数据竞争和逻辑混乱。

**修复**：只在主循环中使用 channel 收集投票结果后判断，移除 goroutine 内部的判断。

#### Bug 3: 缺少锁保护

**问题**：多个地方读取 `rf.Role`, `rf.term` 等共享状态时未加锁。

**修复**：所有访问共享状态的地方都加锁。

#### Bug 4: 单服务器网络分区成为 Leader

**问题**：在网络分区中，只有一个服务器连接时，它不应该成为 leader（因为无法获得多数派）。

**原因**：投票判断条件使用了 `>=` 而不是 `>`。

**修复**：确保使用严格多数判断 `votes > total/2`。

## 测试说明

### 测试用例

#### TestInitialElection3A
- **目的**：验证初始选举
- **步骤**：
  1. 启动 3 个服务器
  2. 检查是否选出了 leader
  3. 检查所有服务器的 term 是否一致
  4. 在没有网络故障的情况下，检查 leader 和 term 是否保持稳定

#### TestReElection3A
- **目的**：验证网络故障后的重新选举
- **步骤**：
  1. 选出初始 leader
  2. 断开 leader 连接，检查是否选出新 leader
  3. 重新连接旧 leader，检查它是否转换为 follower
  4. 断开多数派服务器，检查是否**没有** leader（无法形成多数派）
  5. 恢复多数派连接，检查是否选出新 leader

#### TestManyElections3A
- **目的**：验证多次选举的稳定性
- **步骤**：在 10 秒内随机断开/连接服务器，检查始终只有一个 leader

### 运行测试

```bash
cd src/raft1

# 运行所有 3A 测试
go test -run 3A

# 带竞争检测
go test -race -run 3A

# 运行特定测试
go test -run TestInitialElection3A

# 多次运行检测不稳定性
for i in {1..100}; do
    go test -run TestReElection3A
done
```

### 常见测试失败原因

1. **`checkNoLeader()` 失败**
   - 原因：单服务器网络分区时仍认为自己是 leader
   - 修复：确保投票判断使用 `votes > total/2`

2. **数据竞争**
   - 原因：访问共享状态未加锁
   - 修复：使用 `go test -race` 检测，所有共享状态访问加锁

3. **选举超时/死锁**
   - 原因：锁管理不当，RPC 发送时持有锁
   - 修复：发送 RPC 前释放锁

4. **Term 不一致**
   - 原因：收到更高 term 时未及时更新
   - 修复：在所有 RPC handler 中检查 term 并更新

## 性能优化建议

### 选举超时时间
- 论文建议：150-300ms
- 测试要求：1 秒内完成选举
- 调优：
  - 太短：频繁选举，浪费资源
  - 太长：故障恢复慢

### 心跳间隔
- 建议：选举超时的 1/3 到 1/5
- 例如：选举超时 500ms，心跳间隔 150ms

### 并发度
- 并发发送 RequestVote RPC（使用 goroutine）
- 并发发送心跳（每个 follower 一个 goroutine）

## 调试技巧

### 日志
```go
log.Printf("server %d in term %d becomes leader", rf.me, rf.term)
log.Printf("server %d received vote from server %d", rf.me, voter)
```

### 打印状态
```go
func (rf *Raft) String() string {
    rf.mu.Lock()
    defer rf.mu.Unlock()
    return fmt.Sprintf("Server %d: role=%s term=%d", rf.me, rf.Role, rf.term)
}
```

### 使用 DPrintf
```go
const Debug = true

func DPrintf(format string, a ...interface{}) {
    if Debug {
        log.Printf(format, a...)
    }
}
```

### 竞争检测
```bash
go test -race -run TestReElection3A
```

### 可视化工具
- 使用课程提供的 log 解析工具
- 绘制时间线图，观察选举过程

## 下一步：Lab 3B

Lab 3B 将实现：
- 日志复制（Log Replication）
- 日志一致性检查
- 持久化状态

## 参考资料

- [Raft 论文](https://raft.github.io/raft.pdf)
- [Raft 可视化](http://thesecretlivesofdata.com/raft/)
- [MIT 6.5840 课程网站](https://pdos.csail.mit.edu/6.824/)

## 总结

Lab 3A 实现了 Raft 的核心选举机制，关键要点：

1. **严格多数**：`votes > n/2` 才能成为 leader
2. **锁管理**：所有共享状态访问必须加锁
3. **Term 管理**：始终维护最新的 term
4. **心跳机制**：定期发送心跳维持 leader 身份
5. **并发安全**：通过 `-race` 检测数据竞争

通过正确实现这些机制，可以在网络分区、服务器故障等情况下保证选举的正确性和一致性。

