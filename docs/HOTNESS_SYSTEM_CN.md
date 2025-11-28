# 热度排序系统

## 概述

热度排序系统根据投票和评论计算并维护帖子的热度分数。它支持两种流行的排序算法：Reddit 热度排序和 Hacker News 排序。

## 架构

### 组件

1. **HotnessService**: 使用可配置算法计算热度分数
2. **HotnessWorker**: 监听投票和评论事件并触发重新计算
3. **WorkerManager**: 管理工作者生命周期和消息队列订阅

### 算法

#### Reddit 算法

公式: `log10(max(|score|, 1)) + sign(score) * seconds / 45000`

其中：
- `score = 赞成票 - 反对票`
- `seconds = 自纪元以来的秒数`
- `45000 ≈ 12.5 小时（秒）`

**特点：**
- 时间加权：较新的帖子获得提升
- 对数投票缩放：更多投票的边际效益递减
- 同时考虑正面和负面投票

#### Hacker News 算法

公式: `(score - 1) / (age + 2)^gravity`

其中：
- `score = 赞成票 - 反对票 + 1`
- `age = 帖子创建后的小时数`
- `gravity = 1.8`（控制衰减率）

**特点：**
- 强时间衰减：旧帖子快速下降
- 线性投票缩放（在分子内）
- 重力因子控制帖子老化速度

## 配置

在环境变量或配置文件中设置算法：

```bash
HOTNESS_ALGORITHM=reddit  # 或 "hackernews"
```

默认值: `reddit`

## 使用方法

### 服务初始化

```go
import (
    "github.com/kobayashirei/airy/internal/service"
    "github.com/kobayashirei/airy/internal/repository"
)

// 创建热度服务
hotnessService := service.NewHotnessService(
    postRepo,
    entityCountRepo,
    service.AlgorithmReddit, // 或 service.AlgorithmHackerNews
)

// 计算帖子热度
score, err := hotnessService.CalculateHotness(ctx, post, counts)

// 重新计算并更新热度
newScore, err := hotnessService.RecalculatePostHotness(ctx, postID)
```

### 工作者设置

```go
import (
    "github.com/kobayashirei/airy/internal/service"
    "github.com/kobayashirei/airy/internal/mq"
)

// 创建热度工作者
hotnessWorker := service.NewHotnessWorker(
    hotnessService,
    searchClient,
)

// 创建工作者管理器
workerManager := service.NewWorkerManager(
    messageQueue,
    hotnessWorker,
    logger,
)

// 启动所有工作者
if err := workerManager.Start(); err != nil {
    log.Fatal(err)
}

// 关闭时停止工作者
defer workerManager.Stop()
```

## 事件流程

1. 用户对帖子投票或创建评论
2. 投票/评论服务发布事件到消息队列
3. HotnessWorker 接收事件
4. 工作者触发热度重新计算
5. 新分数保存到数据库
6. Elasticsearch 索引更新新分数

## 消息队列主题

热度工作者订阅：
- `vote.created` - 帖子新投票
- `vote.updated` - 投票更改（赞成变反对或反之）
- `vote.deleted` - 投票删除
- `comment.created` - 帖子新评论
- `comment.deleted` - 评论删除

## 数据库更新

系统更新两个位置：
1. **posts.hotness_score** - 主数据库字段
2. **Elasticsearch posts 索引** - 用于搜索和排序

更新是原子的，使用专用的仓库方法避免竞态条件。

## 性能考虑

### 异步处理
- 热度重新计算通过消息队列异步进行
- 不阻塞用户请求
- 最终一致性模型

### 缓存
- 热度更新时不会使帖子缓存失效
- 热度主要用于排序，而非显示
- 缓存失效发生在内容更新时

### Elasticsearch 同步
- ES 更新是尽力而为
- 失败会被记录但不会导致操作失败
- ES 可以与数据库最终一致

## 测试

运行热度服务测试：

```bash
go test -v ./internal/service -run TestCalculate
```

## 监控

需要监控的关键指标：
- 热度计算延迟
- 消息队列处理速率
- ES 同步成功率
- 工作者错误率

## 故障排除

### 帖子未出现在热门 Feed 中
1. 检查热度工作者是否运行
2. 验证消息队列连接
3. 检查工作者日志中的错误
4. 验证帖子是否有投票或评论

### 热度分数似乎不正确
1. 验证算法配置
2. 检查帖子时间戳（published_at vs created_at）
3. 验证 entity_counts 表数据正确
4. 检查算法参数

### ES 索引不同步
1. 检查 ES 连接
2. 查看工作者日志中的 ES 错误
3. 如需要考虑手动重建索引
4. 验证 ES 映射包含 hotness_score 字段

## 未来增强

潜在改进：
- 通过配置自定义算法参数
- 所有帖子的定时批量重新计算
- 算法 A/B 测试支持
- 衰减因子配置
- 评论权重配置
