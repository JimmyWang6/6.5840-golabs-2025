# 6.5840 Go Labs 总览（Lab1 & Lab2）

A Go implementation of core labs from MIT 6.5840 (distributed systems), including a MapReduce framework and a versioned key/value server with a lock service.

当前仓库主要包含：

- **Lab1：MapReduce 分布式计算框架**
- **Lab2：版本化 Key/Value 存储与基于 KV 的分布式锁**

代码主要位于 `src/` 目录，根目录下的 `Lab1-MapReduce-Design.md` 和 `Lab2-KVServer-Design.md` 是这两个实验的详细设计文档。

---

## 环境与构建

- **语言与环境**
  - Go 版本：以 `src/go.mod` 为准（通常为 Go 1.20+）
  - OS：Linux / macOS（开发和测试在 macOS 上完成）

- **基础使用方式**
  - 进入源码目录：

    ```bash
    cd src
    ```

  - 按子目录运行各实验测试，例如：

    ```bash
    # 运行 MapReduce 测试
    go test ./mr

    # 运行 Lab2 KV Server 相关测试
    go test ./kvsrv1
    ```

  - MapReduce 也可以使用提供的脚本（在 `src/main/` 下）：

    ```bash
    cd src/main
    ./test-mr.sh
    ./test-mr-many.sh
    ```

> 实际测试命令可以按课程说明或本地需求再补充到本 README 中。

---

## Lab1：MapReduce 框架

### 目标

- 实现一个简化版的 **MapReduce** 系统，包含：
  - 单个 `Coordinator`（协调者）
  - 多个 `Worker`（工作进程）
- 理解并实践：
  - Map/Reduce 两阶段执行模型
  - 任务划分与调度
  - Worker 故障检测与任务重分配
  - 基于 RPC 的进程间通信

### 核心设计与架构

**主要目录与文件：**

- `src/mr/`
  - `coordinator.go`：Coordinator 实现
  - `worker.go`：Worker 实现
  - `rpc.go`：Coordinator 与 Worker 的 RPC 接口定义
  - `coordinator_test.go` 等测试代码
- `src/mrapps/`：若干 MapReduce 应用（`wc.go`, `indexer.go`, `jobcount.go` 等）
- `src/main/`
  - `mrcoordinator.go` / `mrworker.go` / `mrsequential.go`
  - 测试脚本 `test-mr.sh`, `test-mr-many.sh`
  - 测试输入数据 `pg-*.txt` 等

**系统角色：**

- **Coordinator (`coordinator.go`)**
  - 负责：
    - 管理 Map/Reduce 所有任务的状态
    - 为 Worker 分配任务
    - 追踪 Worker 状态和心跳，处理失败与任务重分配
    - 管理全局执行阶段：`map` → `reduce` → `finish`
  - 关键字段（简化）：
    - `State`：当前阶段（"map" / "reduce" / "finish"）
    - `Files []File`：所有任务列表（包括 map 和 reduce）
    - `Workers map[int]*WorkerItem`：已注册 Worker 集合，含心跳时间等
    - `NReduce`：reduce 任务数
  - 任务结构 `File`：
    - `Name`：文件名或 reduce 任务标识
    - `Index`：任务索引
    - `Assigned` / `Done`：是否已分配 / 已完成
    - `Owner`：当前执行该任务的 Worker ID
    - `Type`：任务类型（`map` 或 `reduce`）

- **Worker (`worker.go`)**
  - 启动后：
    - 向 Coordinator `Register`，获取 Worker ID、`NReduce` 等配置信息
    - 进入循环：
      1. 报告上一个任务完成情况（如有），通过 `Report` 请求新任务
      2. 执行分配到的 map 或 reduce 任务
      3. 周期性发送 `Heartbeat` 通知 Coordinator 自己仍然存活
  - Map 任务：
    - 读取输入文件，调用用户定义的 `mapf` 生成 `(key, value)` 对
    - 按 `key` 的哈希分区，把中间结果写入 `mr-<mapTask>-<reduceTask>` 形式的文件
  - Reduce 任务：
    - 收集中间文件 `mr-*-<reduceTask>`，按 `key` 分组排序
    - 调用用户定义的 `reducef` 生成最终结果，写入 `mr-out-<reduceTask>`

**执行阶段与调度逻辑：**

- **Map 阶段**
  - Coordinator 将每个输入文件对应为一个 map 任务（`Type = map`）
  - 所有 map 任务完成后，才会进入 reduce 阶段
- **Reduce 阶段**
  - Coordinator 为每个 reduce 分区创建 reduce 任务（`Type = reduce`）
  - 所有 reduce 任务完成后，进入 `finish` 状态
- **任务分配**
  - Worker 通过 `Report` 请求任务：
    - 若有未分配且未完成的任务，则标记为 Assigned 并返回给 Worker
    - 若当前阶段任务都已完成，Coordinator 切换到下一阶段
    - 若全部完成，返回 `done`，Worker 退出
- **故障检测与容错**
  - Coordinator 为每个 Worker 维护 `LastSeen` 时间戳
  - 后台协程周期性扫描：
    - 若某 Worker 超过阈值未发送心跳，认为已死亡
    - 把该 Worker 已分配但未完成的任务标记为未分配，交给其他 Worker 重试

### 模式与特点

- 采用简单的全局互斥锁保证 Coordinator 内部数据一致性
- 任务分配策略是“先到先得”的队列式调度，便于实现，也具备一定负载均衡
- Map 和 Reduce 阶段严格分离，确保与课程测试要求匹配
- 中间文件命名和 JSON 编码格式与课程规范保持一致

### 运行与测试

示例（可按需调整）：

```bash
cd src

# 运行 mr 包自带测试
go test ./mr

# 使用脚本运行完整 MapReduce 测试
cd main
./test-mr.sh
./test-mr-many.sh
```

---

## Lab2：版本化 Key/Value Server 与锁服务

> 注：具体文件结构可根据最终实现情况略有出入，这里以本仓库现有目录为参考，如 `src/kvsrv1`, `src/kvraft1`, `src/kvtest1`, `src/models1` 等。

### 目标

- 实现一个支持**版本号**的 Key/Value 存储服务：
  - 使用版本实现 **乐观并发控制**（Optimistic Concurrency Control）
- 在**不可靠网络**环境下，通过客户端重试逻辑实现接近 *at-most-once* 的语义：
  - 使用特殊错误码 `ErrMaybe` 表示操作结果不确定
- 基于版本化 KV，构建一个**分布式锁服务**：
  - 利用 KV + 版本号 + `ErrMaybe` 来安全实现 `Acquire` / `Release`

### 核心设计与组件

**主要目录与文件（示例）：**

- `src/kvsrv1/`
  - `server.go`：版本化 KV Server 实现
  - `client.go`：KV 客户端封装（Clerk）
  - `kvsrv_test.go`：基础与并发测试
  - `lock/`：基于 KV 的锁服务实现
  - `rpc/`：RPC 参数与错误码定义
- `src/kvtest1/kvtest.go`、`porcupine.go`：线性一致性测试工具
- `src/models1/kv.go`：KV 操作的模型定义（供 Porcupine 使用）

#### KV Server：版本化存储与乐观并发

- 服务器内部状态示意：

  ```go
  type KVServer struct {
      mu sync.Mutex
      kv map[string]PutArgs // 或类似结构，包含 Value 和 Version
  }
  ```

- **Get** 逻辑：
  - 若 key 存在：返回当前 `value` 和 `version`，`Err = OK`
  - 若 key 不存在：`Err = ErrNoKey`
- **Put** 逻辑（简化规则）：
  - 若 key 已存在：
    - `args.Version == current.Version`：允许更新，写入新值并将版本号加一，`Err = OK`
    - 否则：`Err = ErrVersion`（并发冲突）
  - 若 key 不存在：
    - `args.Version == 0`：创建新 key，版本设为 1，`Err = OK`
    - 否则：`Err = ErrNoKey`
- **并发控制：**
  - `Get` / `Put` 在服务器内受 `mu` 保护：
    - 保证检查版本与更新写入是原子操作
    - 实现简单、易于验证正确性

#### 客户端：重试与 ErrMaybe 语义

- 客户端封装 `Get` 和 `Put`：
  - **Get**
    - 幂等：RPC 失败时可简单重试直到成功
  - **Put**
    - 具有副作用：若第一次 RPC 成功但响应丢失，重试可能导致“重复提交”问题
    - 客户端策略：
      - 记录是否发生过超时/错误并重试
      - 若在重试后收到 `ErrVersion` 等错误，无法判断第一次是否已生效 → 返回 `ErrMaybe`
- **错误码：**
  - `OK`：操作成功
  - `ErrNoKey`：key 不存在
  - `ErrVersion`：版本冲突
  - `ErrMaybe`：由于网络不可靠和重试，客户端本地无法判断 Put 是否成功

#### 基于 KV 的分布式锁

- 锁状态编码：
  - 用特定 key 表示一把锁（如 `L`）
  - 值为空串 `""`：锁空闲
  - 值为某客户端标识：锁被该客户端持有
- **Acquire(L)** 大致算法：
  1. 调用 `Get(L)` 读取当前值与版本 `v`
  2. 若值不为空 → 锁被占用，等待并重试
  3. 若值为空：
     - 调用 `Put(L, name, v)` 试图占用锁
       - `OK`：成功获取锁
       - `ErrVersion`：被其他客户端抢先，重试
       - `ErrMaybe`：
         - 再次 `Get(L)`：
           - 若值为 `name`，则认为成功获取锁
           - 否则认为失败，继续重试
- **Release(L)** 大致算法：
  1. 使用持有锁时记录的版本 `v`，调用 `Put(L, "", v+1)` 释放锁
     - `OK`：释放成功
     - `ErrMaybe`：
       - `Get(L)` 检查：
         - 若值为空串，视为已释放
         - 否则继续重试
  2. 理论上不应出现 `ErrVersion` / `ErrNoKey`，若出现通常表示逻辑使用错误

### 并发与线性一致性

- **服务器端**
  - 单互斥锁串行化对 `kv` map 的访问，保证每次 Get/Put 都是原子的
- **线性一致性检查**
  - 使用 `kvtest1/porcupine.go` 等工具记录所有操作历史
  - 利用 Porcupine 检查是否存在与规范一致的线性化执行顺序
- **测试要点**
  - 多客户端对同一 key 并发 Put：检查版本号和最终值
  - 模拟 RPC 超时/丢包：检查 `ErrMaybe` 是否正确出现和被正确处理
  - 大量操作下的内存占用与性能（如 `TestMemPutManyClientsReliable` 等）

### 已实现能力摘要（Lab2）

- 版本化单机 KV 存储：
  - 提供基于版本号的乐观并发控制
- 客户端封装与 `ErrMaybe` 语义：
  - 在不可靠网络下仍能给出合理的一致性保证
- 基于 KV 的分布式锁：
  - 使用 Get/Put + version + ErrMaybe，实现安全的锁获取与释放
- 支持 Porcupine 线性一致性验证的测试框架

---

## 仓库结构导览

- 根目录
  - `Lab1-MapReduce-Design.md`：Lab1 详细设计文档
  - `Lab2-KVServer-Design.md`：Lab2 详细设计文档
  - `Makefile`：可能用于统一构建/测试
- `src/`
  - `mr/`：MapReduce 核心实现（Coordinator / Worker / RPC / 测试）
  - `mrapps/`：MapReduce 应用示例（wc, indexer, jobcount 等）
  - `main/`：可执行程序入口与 MR 测试脚本
  - `kvsrv1/`：Lab2 版本化 KV Server、Client、Lock 实现及测试
  - `kvtest1/`：KV 测试工具与 Porcupine 线性一致性检查
  - `models1/`：KV 操作模型定义
  - `raft1/`, `shardkv1/` 等：后续实验目录（可逐步实现）

---

## 后续工作与扩展方向

- **MapReduce（Lab1）**
  - 实现备份任务（backup tasks）减少慢任务影响
  - 任务调度中考虑数据本地性、Worker 历史性能等
  - 支持 Combiner 函数，减少中间数据规模
- **KV Server & Lock（Lab2）**
  - 支持更多操作（Append、PutIf、事务等）
  - 在单机版本化 KV 上叠加复制和分片（对应后续 Raft / ShardKV 实验）
  - 优化锁服务（租约、超时、可重入等）

