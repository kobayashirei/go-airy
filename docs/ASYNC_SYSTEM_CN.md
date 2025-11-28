# 异步任务系统

本文档描述 Airy 应用中的异步任务处理系统，由两个主要组件组成：任务池和消息队列。

## 概述

异步系统使应用能够：
- 处理计算密集型任务而不阻塞 HTTP 请求
- 异步处理 IO 密集型操作
- 通过事件驱动架构解耦组件
- 通过在多个工作者之间分配工作来水平扩展
- 提高响应时间和系统吞吐量

## 架构

```
┌─────────────┐
│   Handler   │
└──────┬──────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼
┌─────────────┐  ┌──────────────┐
│   任务池    │  │   消息队列   │
└──────┬──────┘  └──────┬───────┘
       │                │
       ▼                ▼
┌─────────────┐  ┌──────────────┐
│   Worker    │  │   订阅者     │
└─────────────┘  └──────────────┘
```

## 组件

### 1. 任务池 (ants)

任务池管理一个协程池，用于高效执行任务。

**位置**: `internal/taskpool/`

**主要特性**:
- 固定大小的协程池防止资源耗尽
- 带错误处理的任务提交
- Panic 恢复
- 指标（运行中、空闲、等待中的协程）
- 带超时的优雅关闭
- 上下文取消支持

**使用场景**:
- CPU 密集型计算
- 短期任务（< 5 分钟）
- 不需要持久化的任务
- 应在服务器关闭前完成的任务

**示例**:
```go
pool, _ := taskpool.NewPool(&taskpool.Config{
    Size: 1000,
    Logger: logger,
})
defer pool.Release()

// 提交任务
pool.SubmitFunc(func(ctx context.Context) error {
    // 执行一些工作
    return nil
})
```

### 2. 消息队列 (RabbitMQ)

消息队列提供组件之间可靠、持久的消息传递。

**位置**: `internal/mq/`

**主要特性**:
- 基于主题的路由
- 持久消息传递
- 自动重连
- 每个主题多个订阅者
- 消息确认和重新入队
- JSON 序列化

**使用场景**:
- 长时间运行的任务
- 必须在服务器重启后存活的任务
- 需要分布在多个服务器上的任务
- 事件驱动的工作流
- 需要保证传递的任务

**示例**:
```go
mq, _ := mq.NewRabbitMQ(&mq.Config{
    URL: "amqp://guest:guest@localhost:5672/",
    ExchangeName: "airy.events",
    Logger: logger,
})
defer mq.Close()

// 发布事件
publisher := mq.NewPublisher(mq)
publisher.PublishPostPublished(ctx, postID, authorID, nil, title)

// 订阅事件
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    // 处理事件
    return nil
})
```

## 使用场景

### 帖子发布流程

当用户发布帖子时，会触发多个异步任务：

```go
// 在 PostService.Create() 中
func (s *PostService) Create(ctx context.Context, post *Post) error {
    // 1. 保存帖子到数据库
    if err := s.repo.Create(ctx, post); err != nil {
        return err
    }

    // 2. 发布事件到消息队列
    s.publisher.PublishPostPublished(ctx, post.ID, post.AuthorID, post.CircleID, post.Title)

    return nil
}

// 订阅者异步处理事件：

// 订阅者 1: 更新搜索索引
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    // 提交到任务池处理
    return pool.SubmitFunc(func(ctx context.Context) error {
        return searchService.IndexPost(ctx, event.PostID)
    })
})

// 订阅者 2: 更新用户 Feed
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    return pool.SubmitFunc(func(ctx context.Context) error {
        return feedService.PushToFollowers(ctx, event.PostID, event.AuthorID)
    })
})

// 订阅者 3: 发送通知
mq.Subscribe(mq.TopicPostPublished, func(ctx context.Context, msg []byte) error {
    var event mq.PostPublishedEvent
    json.Unmarshal(msg, &event)
    
    return pool.SubmitFunc(func(ctx context.Context) error {
        return notificationService.NotifyFollowers(ctx, event.PostID, event.AuthorID)
    })
})
```

### 评论创建流程

```go
// 在 CommentService.Create() 中
func (s *CommentService) Create(ctx context.Context, comment *Comment) error {
    // 1. 保存评论到数据库
    if err := s.repo.Create(ctx, comment); err != nil {
        return err
    }

    // 2. 发布事件
    s.publisher.PublishCommentCreated(
        ctx,
        comment.ID,
        comment.PostID,
        comment.AuthorID,
        comment.ParentID,
        comment.Content,
    )

    return nil
}

// 订阅者：
// - 更新帖子评论计数
// - 通知帖子作者
// - 通知父评论作者
// - 解析 @提及并通知被提及的用户
```

### 投票处理流程

```go
// 在 VoteService.Vote() 中
func (s *VoteService) Vote(ctx context.Context, vote *Vote) error {
    // 1. 创建或更新投票（幂等）
    if err := s.repo.Upsert(ctx, vote); err != nil {
        return err
    }

    // 2. 发布事件
    s.publisher.PublishVoteCreated(
        ctx,
        vote.ID,
        vote.UserID,
        vote.EntityType,
        vote.EntityID,
        vote.VoteType,
    )

    return nil
}

// 订阅者：
// - 更新 entity_counts 表中的投票计数
// - 重新计算热度分数
// - 通知内容作者（如果不是自己投票）
```

## 事件类型

### 帖子事件
- `post.published` - 帖子发布
- `post.updated` - 帖子更新
- `post.deleted` - 帖子删除
- `post.voted` - 帖子收到投票

### 评论事件
- `comment.created` - 评论创建
- `comment.deleted` - 评论删除
- `comment.voted` - 评论收到投票

### 用户事件
- `user.followed` - 用户关注另一个用户
- `user.unfollowed` - 用户取消关注
- `user.registered` - 新用户注册

### 圈子事件
- `circle.joined` - 用户加入圈子
- `circle.left` - 用户离开圈子

### 投票事件
- `vote.created` - 投票创建
- `vote.updated` - 投票更新
- `vote.deleted` - 投票删除

## 配置

### 任务池配置

环境变量：
- `GOROUTINE_POOL_SIZE` - 最大协程数（默认: 10000）

代码配置：
```go
config := &taskpool.Config{
    Size:             10000,              // 池大小
    ExpiryDuration:   10 * time.Second,   // 协程过期时间
    PreAlloc:         false,              // 预分配协程
    MaxBlockingTasks: 0,                  // 最大阻塞任务数（0 = 无限）
    Nonblocking:      false,              // 非阻塞模式
    Logger:           logger,             // 日志实例
}
```

### 消息队列配置

环境变量：
- `MQ_HOST` - RabbitMQ 主机（默认: localhost）
- `MQ_PORT` - RabbitMQ 端口（默认: 5672）
- `MQ_USER` - RabbitMQ 用户（默认: guest）
- `MQ_PASSWORD` - RabbitMQ 密码（默认: guest）

代码配置：
```go
config := &mq.Config{
    URL:          "amqp://guest:guest@localhost:5672/",
    ExchangeName: "airy.events",
    Logger:       logger,
}
```

## 最佳实践

### 任务池

1. **任务大小**: 保持任务小而专注（< 5 分钟）
2. **错误处理**: 始终从任务返回错误以便记录
3. **上下文**: 在长时间运行的任务中尊重上下文取消
4. **资源清理**: 在任务中使用 defer 进行清理
5. **池大小**: 根据 CPU 核心数和工作负载调整池大小

### 消息队列

1. **幂等性**: 设计处理器为幂等的（消息可能被多次传递）
2. **错误处理**: 返回错误以触发消息重新入队
3. **超时**: 使用上下文超时防止阻塞
4. **事件模式**: 保持事件向后兼容
5. **日志**: 记录所有事件处理以便调试
6. **死信队列**: 为失败消息配置 DLQ

### 组合使用

1. **关注点分离**: 使用 MQ 进行分发，任务池进行执行
2. **背压**: 任务池为 MQ 消费者提供自然的背压
3. **监控**: 监控池指标和队列深度
4. **优雅关闭**: 在关闭 MQ 连接前等待任务池

## 监控

### 任务池指标

```go
pool.Running()  // 运行中的协程数
pool.Free()     // 可用的协程数
pool.Waiting()  // 等待中的任务数
pool.Cap()      // 池容量
```

### 消息队列指标

通过 RabbitMQ 管理界面监控：
- 队列深度
- 消息速率（发布/传递）
- 消费者数量
- 未确认消息

## 错误处理

### 任务池错误

- 任务错误会被记录但不会停止池
- Panic 会被恢复并记录
- 失败的任务不会自动重试

### 消息队列错误

- 处理器错误触发消息重新入队
- 连接错误触发自动重连
- 失败的消息可以路由到死信队列

## 测试

### 任务池测试

```bash
go test ./internal/taskpool/...
```

### 消息队列测试

```bash
# 启动 RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 rabbitmq:3-management

# 运行测试
go test ./internal/mq/...
```

## 示例

参见 `examples/async_example.go` 获取完整的工作示例：
- 单独使用任务池
- 单独使用消息队列
- 组合使用任务池和消息队列

## 故障排除

### 任务池问题

**问题**: 任务不执行
- 检查池是否已关闭
- 检查池大小是否足够
- 检查任务中是否有 panic

**问题**: 内存使用过高
- 减小池大小
- 检查协程泄漏
- 确保任务及时完成

### 消息队列问题

**问题**: 消息未传递
- 检查 RabbitMQ 是否运行
- 检查连接 URL 是否正确
- 检查交换机和队列绑定

**问题**: 消息反复重新入队
- 检查处理器是否不必要地返回错误
- 检查是否有无限错误循环
- 配置死信队列

**问题**: 连接断开
- 检查网络稳定性
- 检查 RabbitMQ 资源限制
- 监控重连尝试

## 未来增强

- [ ] 支持 Kafka 作为替代消息队列
- [ ] 消息优先级支持
- [ ] 延迟消息传递
- [ ] 消息批处理
- [ ] 事件重放功能
- [ ] 事件模式验证
- [ ] 分布式追踪集成
- [ ] 指标导出到 Prometheus
