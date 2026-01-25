# Redis缓存实现方案

## 架构设计

### 缓存策略

**Cache-Aside + 懒加载 + 写操作删除缓存**

- 读操作：先查缓存，未命中则查DB并写缓存
- 写操作：更新DB后删除相关缓存
- Sorted Set按需重建（懒加载）
- 并发安全：Redis分布式锁（10秒超时，3次重试）

### 数据结构

| 类型 | 用途 | Key格式 | TTL |
|------|------|---------|-----|
| **Hash** | 单个对象详情 | `cache:todo:{id}` | 1小时 |
| **Hash** | 单个Tag | `cache:tag:{id}` | 30分钟 |
| **Sorted Set** | Todo列表（基础） | `cache:todos:user:{id}:sorted:due_date:asc` | 10分钟 |
| **Sorted Set** | Todo列表（状态过滤） | `cache:todos:user:{id}:sorted:status:{status}:due_date:asc` | 10分钟 |
| **Sorted Set** | Todo列表（优先级过滤） | `cache:todos:user:{id}:sorted:priority:high:due_date:asc` | 10分钟 |
| **String** | Todo复杂查询 | `cache:todos:user:{id}:query:{md5}` | 5分钟 |
| **String** | Tags列表 | `cache:tags:page:{page}:limit:{limit}` | 30分钟 |
| **String** | 用户Tags统计 | `cache:tags:my-tags:{id}` | 20分钟 |

---

## Redis数据结构设计

### 1. Hash缓存（单个对象）

#### 1.1 Todo对象
\`\`\`
Key: cache:todo:{id}
Type: Hash
TTL: 3600s (1小时)
\`\`\`

**Hash Fields**：
\`\`\`go
{
  "id": int64,
  "user_id": int64,
  "title": string,
  "description": string,
  "status": string,          // "not_started", "in_progress", "completed"
  "priority": string,        // "low", "medium", "high"
  "due_date": int64,         // Unix timestamp
  "created_at": int64,      // Unix timestamp
  "updated_at": int64       // Unix timestamp
}
\`\`\`

#### 1.2 Tag对象
\`\`\`
Key: cache:tag:{id}
Type: Hash
TTL: 1800s (30分钟)
\`\`\`

### 2. Sorted Set缓存（Todo列表）

#### 2.1 基础Sorted Set（4个）
\`\`\`
cache:todos:user:{userID}:sorted:due_date:asc
cache:todos:user:{userID}:sorted:due_date:desc
cache:todos:user:{userID}:sorted:created_at:desc
cache:todos:user:{userID}:sorted:title:asc
\`\`\`

**Score规则**：
- \`\`\`due_date:asc\`\`\`: \`\`\`score = due_date.Unix()\`\`\`
- \`\`\`due_date:desc\`\`\`: \`\`\`score = -due_date.Unix()\`\`\`
- \`\`\`created_at:desc\`\`\`: \`\`\`score = -created_at.Unix()\`\`\`
- \`\`\`title:asc\`\`\`: \`\`\`score = FNV32Hash(title)\`\`\`

**无截止日期处理**：
- \`\`\`due_date:asc\`\`\`: \`\`\`score = +∞\`\`\`（排在最后）
- \`\`\`due_date:desc\`\`\`: \`\`\`score = -∞\`\`\`（排在最前）

#### 2.2 状态过滤Sorted Set（3个）
\`\`\`
cache:todos:user:{userID}:sorted:status:not_started:due_date:asc
cache:todos:user:{userID}:sorted:status:in_progress:due_date:asc
cache:todos:user:{userID}:sorted:status:completed:due_date:asc
\`\`\`

#### 2.3 优先级过滤Sorted Set（1个）
\`\`\`
cache:todos:user:{userID}:sorted:priority:high:due_date:asc
\`\`\`

**总计：8个Sorted Set**

### 3. String缓存（复杂查询+Tags）

#### 3.1 Todo复杂查询
\`\`\`
Key: cache:todos:user:{userID}:query:{md5(params)}
Type: String
TTL: 300s (5分钟)
Value: JSON序列化的完整响应
\`\`\`

**使用场景**：
- 日期范围过滤：\`\`\`?due_date_from=2024-01-01&due_date_to=2024-01-31\`\`\`
- 搜索功能：\`\`\`?search=关键词\`\`\`
- 多条件组合：\`\`\`?status=in_progress&priority=high\`\`\`

#### 3.2 Tags列表
\`\`\`
Key: cache:tags:page:{page}:limit:{limit}
Type: String
TTL: 1800s (30分钟)
\`\`\`

#### 3.3 用户Tags统计
\`\`\`
Key: cache:tags:my-tags:{userID}
Type: String
TTL: 1200s (20分钟)
\`\`\`

---

## 缓存读写流程

### 读取流程

#### 流程1：基础查询（使用Sorted Set）
\`\`\`
1. 构建Sorted Set key
   ↓
2. 检查Sorted Set是否存在
   ├─ 存在 → 继续
   └─ 不存在 → 从DB加载用户所有Todo → 重建Sorted Set
   ↓
3. ZRANGE获取ID列表（支持分页）
   ↓
4. 批量HGETALL从Hash获取Todo详情
   ↓
5. 组装Response返回
\`\`\`

#### 流程2：复杂查询（使用String缓存）
\`\`\`
1. 构建查询参数的MD5 hash
   ↓
2. 尝试从缓存获取
   ├─ 命中 → 直接返回JSON
   └─ 未命中 → 继续下步
   ↓
3. 从DB查询（含分页、过滤、排序）
   ↓
4. 缓存完整响应（JSON）
   ↓
5. 返回
\`\`\`

### 写入流程

#### 流程1：创建Todo
\`\`\`
1. 获取分布式锁（todo:user:{userID}）
   ↓
2. 写入DB
   ↓
3. Pipeline操作：
   - HSET缓存: cache:todo:{id}
   - ZADD到所有8个Sorted Set（如果存在）
   - EXPIRE设置TTL
   ↓
4. DEL所有查询缓存：cache:todos:user:{userID}:query:*
   ↓
5. EXEC Pipeline
   ↓
6. 释放锁
\`\`\`

#### 流程2：更新Todo
\`\`\`
1. 获取分布式锁（todo:user:{userID}）
   ↓
2. 写入DB
   ↓
3. Pipeline操作：
   - HSET更新Hash: cache:todo:{id}
   - ZREM从所有Sorted Set移除旧ID
   - ZADD添加新ID到所有Sorted Set
   - EXPIRE设置TTL
   ↓
4. DEL所有查询缓存
   ↓
5. EXEC Pipeline
   ↓
6. 释放锁
\`\`\`

#### 流程3：删除Todo
\`\`\`
1. 获取分布式锁（todo:user:{userID}）
   ↓
2. 从DB获取Todo（用于状态处理）
   ↓
3. Pipeline操作：
   - DEL Hash缓存: cache:todo:{id}
   - ZREM从所有8个Sorted Set移除ID
   - DEL所有查询缓存
   ↓
4. EXEC Pipeline
   ↓
5. 释放锁
\`\`\`

#### 流程4：更新Todo状态（特殊处理）
\`\`\`
1. 获取分布式锁（todo:user:{userID}）
   ↓
2. 写入DB
   ↓
3. Pipeline操作：
   - HSET更新status字段: cache:todo:{id}
   - ZREM从旧状态Sorted Set移除
   - ZADD到新状态Sorted Set
   - ZADD更新其他Sorted Set
   - DEL所有查询缓存
   ↓
4. EXEC Pipeline
   ↓
5. 释放锁
\`\`\`

### Tag缓存流程

#### 创建/更新/删除Tag
\`\`\`
1. 写入DB
   ↓
2. DEL所有Tags列表缓存：
   - cache:tags:page:*:limit:*
   - cache:tags:my-tags:*
\`\`\`

---

## 缓存失效策略

### Cache-Aside原则

**核心原则**：写操作只更新DB，删除相关缓存

| 操作 | Hash缓存 | Sorted Set缓存 | String缓存 |
|------|---------|---------------|-----------|
| 创建Todo | 创建 | 更新（尝试） | 删除 |
| 更新Todo | 更新 | 更新（先ZREM后ZADD） | 删除 |
| 删除Todo | 删除 | 删除（ZREM） | 删除 |
| 更新状态 | 更新 | 状态迁移 + 更新 | 删除 |
| 创建Tag | - | - | 删除 |
| 更新Tag | - | - | 删除 |
| 删除Tag | 删除（可选） | - | 删除 |

### 失效时机

1. **主动失效**：写操作时立即删除相关缓存
2. **被动失效**：TTL到期自动删除
3. **惰性失效**：Sorted Set重建时自动覆盖

### 冲突处理

- **分布式锁**：10秒超时 + 3次重试 + 100ms间隔
- **锁粒度**：按用户级别锁定（\`\`\`lock:todo:user:{userID}\`\`\`）
- **死锁避免**：锁资源按字典序获取

---

## 接口缓存映射

### Todos相关

| 接口 | 方法 | 缓存方式 | 缓存Key |
|------|------|---------|----------|
| \`\`\`GET /api/v1/todos/:id\`\`\` | GET | 查Hash | \`\`\`cache:todo:{id}\`\`\` |
| \`\`\`GET /api/v1/todos\`\`\`（基础） | GET | 查Sorted Set | \`\`\`cache:todos:user:{id}:sorted:*\`\`\` |
| \`\`\`GET /api/v1/todos\`\`\`（常见过滤） | GET | 查Sorted Set | \`\`\`cache:todos:user:{id}:sorted:status:*\`\`\`<br/>\`\`\`cache:todos:user:{id}:sorted:priority:high:*\`\`\` |
| \`\`\`GET /api/v1/todos\`\`\`（复杂） | GET | 查String | \`\`\`cache:todos:user:{id}:query:{hash}\`\`\` |
| \`\`\`POST /api/v1/todos\`\`\` | POST | 创建 | Hash + Sorted Set |
| \`\`\`PUT /api/v1/todos/:id\`\`\` | PUT | 更新 | Hash + Sorted Set |
| \`\`\`DELETE /api/v1/todos/:id\`\`\` | DELETE | 删除 | Hash + Sorted Set |
| \`\`\`PATCH /api/v1/todos/:id/status\`\`\` | PATCH | 更新 | Hash + Sorted Set |

### Tags相关

| 接口 | 方法 | 缓存方式 | 缓存Key |
|------|------|---------|----------|
| \`\`\`GET /api/v1/tags/:id\`\`\` | GET | 查String | \`\`\`cache:tag:{id}\`\`\` |
| \`\`\`GET /api/v1/tags\`\`\` | GET | 查String | \`\`\`cache:tags:page:{page}:limit:{limit}\`\`\` |
| \`\`\`GET /api/v1/users/my-tags\`\`\` | GET | 查String | \`\`\`cache:tags:my-tags:{id}\`\`\` |
| \`\`\`POST /api/v1/tags\`\`\` | POST | 删除 | 所有Tags缓存 |
| \`\`\`PUT /api/v1/tags/:id\`\`\` | PUT | 删除 | 所有Tags缓存 |
| \`\`\`DELETE /api/v1/tags/:id\`\`\` | DELETE | 删除 | 所有Tags缓存 |

### 管理员相关

**无缓存** - 所有管理员接口直接查询DB

---

## 配置说明

### config.yaml新增配置

\`\`\`yaml
cache:
  todo:
    hash_ttl: 3600s       # 1小时
    sorted_set_ttl: 600s   # 10分钟
    query_ttl: 300s        # 5分钟
  tag:
    ttl: 1800s             # 30分钟
  lock_timeout: 10s        # 10秒
\`\`\`

### Config结构体

\`\`\`go
type CacheConfig struct {
    Todo CacheTodoConfig
    Tag  CacheTagConfig
    LockTimeout int
}

type CacheTodoConfig struct {
    HashTTL       int
    SortedSetTTL  int
    QueryTTL       int
}

type CacheTagConfig struct {
    TTL int
}
\`\`\`

---

## 文件结构

### 新增文件（4个）

\`\`\`
internal/infrastructure/cache/
├── cache_utils.go      # 工具函数
├── lock.go            # 分布式锁
├── todo_cache.go      # Todo缓存逻辑
└── tag_cache.go       # Tag缓存逻辑
\`\`\`

### 修改文件（5个）

\`\`\`
internal/infrastructure/database/redis/connection.go  # 添加Redis方法
internal/infrastructure/config/config.go            # 添加CacheConfig
internal/usecase/todo_usecase.go                  # 集成TodoCache
internal/usecase/tag_usecase.go                   # 集成TagCache
cmd/api/main.go                                  # 初始化缓存层
configs/config.yaml                               # 添加cache配置
\`\`\`

---

## 关键设计决策

### 1. Sorted Set vs String

| 场景 | 选择 | 理由 |
|--------|------|------|
| 基础查询 | Sorted Set | 支持分页、排序、范围查询 |
| 常见过滤（status/priority） | Sorted Set | 使用频率高，值得维护 |
| 复杂查询 | String | 组合太多，维护成本高 |
| 单个Tag | String | 简单查询，与Todos解耦 |

### 2. 懒加载 vs 预热

| 对比 | 优势 | 缺点 |
|------|------|------|
| 懒加载 | 节省内存，按需重建 | 首次查询慢 |
| 预热 | 读操作快，占用内存 | 写操作延迟高 |

**决策**：懒加载

### 3. 缓存失效策略

| 策略 | 说明 | 选择 |
|------|------|------|
| Write-Through | 同时写DB和缓存 | ❌ 复杂，延迟高 |
| Write-Behind | 先写缓存再写DB | ❌ 一致性难保证 |
| Cache-Aside | 写DB后删除缓存 | ✅ 简单，可靠 |
| Refresh-Ahead | 预测失效主动刷新 | ❌ 实现复杂 |

**决策**：Cache-Aside

### 4. 锁粒度

| 粒度 | 优势 | 缺点 |
|------|------|------|
| 全局锁 | 实现简单 | 并发度低 |
| 用户级锁 | 平衡性好 | ✅ 选择 |
| Todo级锁 | 并发度最高 | 实现复杂 |

**决策**：用户级锁（\`\`\`lock:todo:user:{userID}\`\`\`）

---

## 性能优化

### 1. 批量操作

- **ZRANGE**：一次获取ID列表
- **HGETALL**：批量获取Todo详情
- **Pipeline**：原子性操作

### 2. 异步更新

```go
// 删除Todo时，不等待锁
go func() {
    // 后台删除Hash
    redisClient.Del(ctx, "cache:todo:123")
}()
```

### 3. 缓存预热（可选）

**建议场景**：
- 系统启动时预热高频用户数据
- 定时任务刷新过期缓存

---

## 使用说明

### 启动服务

\`\`\`bash
# 启动Redis和MySQL（使用docker-compose）
docker-compose up -d

# 启动API服务
go run cmd/api/main.go
\`\`\`

### 查看Redis缓存

\`\`\`bash
# 连接Redis
redis-cli -h localhost -p 6379

# 查看Todo Hash
HGETALL cache:todo:1

# 查看Sorted Set
ZRANGE cache:todos:user:1:sorted:due_date:asc 0 19
ZCARD cache:todos:user:1:sorted:due_date:asc

# 查看String缓存
GET cache:tags:page:1:limit:20

# 查看TTL
TTL cache:todo:1
TTL cache:todos:user:1:sorted:due_date:asc
\`\`\`

### 清空缓存

\`\`\`bash
# 清空所有缓存
redis-cli FLUSHDB

# 清空特定模式
redis-cli --scan --pattern "cache:todo:*"
redis-cli --scan --pattern "cache:todos:*"
\`\`\`

---

## 监控和调试

### 缓存命中率

\`\`\`bash
# Redis INFO命令
redis-cli INFO stats

# 查看关键指标
# - keyspace_hits
# - keyspace_misses
# - total_commands
\`\`\`

### 日志建议

**需要添加的日志点**：

\`\`\`go
log.Printf("Cache miss: key=%s", key)
log.Printf("Cache hit: key=%s", key)
log.Printf("Sorted set rebuilt: key=%s, count=%d", key, count)
log.Printf("Cache invalidated: pattern=%s", pattern)
log.Printf("Lock acquired: resource=%s", resource)
log.Printf("Lock timeout: resource=%s", resource)
\`\`\`

---

## 常见问题排查

### 1. 缓存未生效

**检查清单**：
- [ ] Redis连接正常
- [ ] Redis配置正确（host、port、database）
- [ ] 缓存层已正确初始化
- [ ] UseCase中正确调用缓存方法
- [ ] TTL配置是否过短

**排查命令**：
\`\`\`bash
# 查看Redis连接
redis-cli PING

# 查看缓存是否存在
redis-cli EXISTS cache:todo:1

# 查看缓存内容
redis-cli GET cache:todo:1
\`\`\`

### 2. 数据不一致

**可能原因**：
- 缓存未及时删除
- 并发写入导致竞争
- DB更新但缓存未更新

**解决方案**：
- 检查Pipeline是否执行成功
- 增加分布式锁重试次数
- 写操作后立即删除缓存

### 3. 内存占用过高

**优化方案**：
- 缩短TTL
- 清理不使用的缓存
- 使用Hash替代多个String
- 监控Redis内存使用

---

## 扩展建议

### 短期优化

1. **添加监控**：缓存命中率、响应时间
2. **添加指标**：Prometheus + Grafana
3. **优化锁策略**：根据QPS调整超时时间
4. **添加缓存预热**：高频用户数据预加载

### 长期优化

1. **多级缓存**：Redis + 本地缓存
2. **缓存穿透保护**：布隆过滤器
3. **缓存雪崩保护**：随机TTL
4. **缓存击穿保护**：互斥锁

---

## 总结

本缓存方案采用 **Sorted Set + Hash**混合策略，实现了：

### 核心特性

- ✅ **高性能**：基础查询使用Sorted Set，支持分页和排序
- ✅ **高可用**：分布式锁保证并发安全
- ✅ **低延迟**：Pipeline批量操作减少网络往返
- ✅ **易维护**：Cache-Aside原则，写操作简单
- ✅ **可扩展**：模块化设计，易于扩展新缓存类型

### 适用场景

- 读多写少的Todo列表查询
- 相对静态的Tags数据
- 中小规模应用（< 1000用户）

### 性能预期

- **缓存命中响应时间**：< 50ms
- **缓存未命中响应时间**：< 200ms
- **写操作延迟**：< 100ms（含锁等待）
- **内存占用**（单用户）：~65KB

### 文件清单

- 新增：4个缓存文件
- 修改：5个现有文件
- 配置：1个配置文件
- 总计：10个文件改动

---

## 附录

### A. Redis Key命名规范

\`\`\`
cache:{entity}:{id}              # 单个对象Hash
cache:{entity}:user:{id}:sorted:*  # Sorted Set列表
cache:{entity}:user:{id}:query:{hash}  # 复杂查询String
cache:{tags}:page:{page}:limit:{limit}  # Tags分页
cache:{tags}:my-tags:{id}  # 用户Tags统计
lock:{resource}                      # 分布式锁
\`\`\`

### B. TTL配置参考

| 缓存类型 | 开发环境 | 测试环境 | 生产环境 |
|---------|---------|---------|---------|
| Hash（Todo） | 3600s (1h) | 1800s (30m) | 7200s (2h) |
| Sorted Set | 600s (10m) | 300s (5m) | 1800s (30m) |
| 查询缓存 | 300s (5m) | 180s (3m) | 900s (15m) |
| Tags列表 | 1800s (30m) | 900s (15m) | 3600s (1h) |

### C. 性能指标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 缓存命中率 | > 80% | 读操作缓存命中率 |
| 平均响应时间 | < 100ms | 有缓存的查询时间 |
| P95响应时间 | < 200ms | 95%的查询时间 |
| Redis内存使用 | < 80% | Redis内存使用率 |
| 锁等待时间 | < 100ms | 获取分布式锁平均时间 |

---

## 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|---------|
| v1.0 | 2025-01-24 | 初始版本，实现Todo和Tag缓存 |
