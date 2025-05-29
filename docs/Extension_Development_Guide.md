# NCore 扩展系统开发指南

Extension System Development Guide

> 本指南面向需要在 NCore 框架下进行扩展开发的开发人员，详细介绍了扩展系统的架构、开发规范和最佳实践。

## 一、扩展系统概述

### 1.1 核心特性

- **动态加载**: 支持文件模式和内置模式的扩展加载
- **依赖管理**: 强弱依赖支持，自动解析依赖关系
- **生命周期管理**: 完整的初始化、运行和清理流程
- **服务发现**: 基于 Consul 的服务注册与发现
- **事件系统**: 统一的事件处理，支持内存和消息队列
- **gRPC 集成**: 可选的 gRPC 服务支持
- **熔断器**: 内置容错机制
- **热重载**: 运行时插件加载和卸载

### 1.2 扩展分类

```go
// 扩展类型常量
const (
    StatusActive       = "active"        // 运行中
    StatusInactive     = "inactive"      // 未激活
    StatusError        = "error"         // 错误状态
    StatusInitializing = "initializing"  // 初始化中
    StatusMaintenance  = "maintenance"   // 维护中
    StatusDisabled     = "disabled"      // 已禁用
)
```

## 二、扩展开发规范

### 2.1 核心接口实现

所有扩展必须实现 `types.Interface` 接口：

```go
type Interface interface {
    // 核心方法
    Name() string
    Version() string
    Init(*config.Config, ManagerInterface) error
    GetMetadata() Metadata

    // 资源方法
    GetHandlers() Handler
    GetServices() Service

    // 依赖方法
    Dependencies() []string

    // 可选方法接口
    OptionalMethods
}
```

### 2.2 可选方法接口

```go
type OptionalMethods interface {
    // 生命周期方法
    PreInit() error
    PostInit() error
    PreCleanup() error
    Cleanup() error

    // 状态和健康检查
    Status() string

    // 依赖管理
    GetAllDependencies() []DependencyEntry

    // 服务发现
    NeedServiceDiscovery() bool
    GetServiceInfo() *ServiceInfo

    // 事件处理
    GetPublisher() any
    GetSubscriber() any

    // HTTP 路由
    RegisterRoutes(*gin.RouterGroup)
}
```

### 2.3 扩展结构模板

```go
package myext

import (
    "github.com/ncobase/ncore/extension/types"
    "github.com/ncobase/ncore/config"
)

// MyExtension 扩展实现
type MyExtension struct {
    types.OptionalImpl  // 嵌入默认实现
    config   *config.Config
    manager  types.ManagerInterface
    handlers *MyHandlers
    services *MyServices
}

// 核心方法实现
func (m *MyExtension) Name() string {
    return "my-extension"
}

func (m *MyExtension) Version() string {
    return "1.0.0"
}

func (m *MyExtension) Dependencies() []string {
    return []string{"core", "auth"}  // 强依赖
}

func (m *MyExtension) GetAllDependencies() []types.DependencyEntry {
    return []types.DependencyEntry{
        {Name: "core", Type: types.StrongDependency},
        {Name: "auth", Type: types.StrongDependency},
        {Name: "user", Type: types.WeakDependency},  // 弱依赖
    }
}

func (m *MyExtension) Init(conf *config.Config, mgr types.ManagerInterface) error {
    m.config = conf
    m.manager = mgr

    // 初始化处理器和服务
    m.handlers = NewMyHandlers()
    m.services = NewMyServices(conf)

    return nil
}

func (m *MyExtension) GetMetadata() types.Metadata {
    return types.Metadata{
        Name:         m.Name(),
        Version:      m.Version(),
        Description:  "我的扩展模块",
        Type:         "module",
        Group:        "business",
        Dependencies: m.Dependencies(),
    }
}

func (m *MyExtension) GetHandlers() types.Handler {
    return m.handlers
}

func (m *MyExtension) GetServices() types.Service {
    return m.services
}

// 可选方法实现
func (m *MyExtension) PostInit() error {
    // 获取弱依赖服务
    if userService, err := m.manager.GetServiceByName("user"); err == nil {
        m.services.SetUserService(userService)
    }
    return nil
}

func (m *MyExtension) RegisterRoutes(router *gin.RouterGroup) {
    // 注册 HTTP 路由
    api := router.Group("/myext")
    {
        api.GET("/status", m.handlers.GetStatus)
        api.POST("/action", m.handlers.DoAction)
    }
}

func (m *MyExtension) NeedServiceDiscovery() bool {
    return true
}

func (m *MyExtension) GetServiceInfo() *types.ServiceInfo {
    return &types.ServiceInfo{
        Address: "localhost:8080",
        Tags:    []string{"api", "v1"},
        Meta:    map[string]string{"version": m.Version()},
    }
}
```

## 三、扩展注册机制

### 3.1 自动注册（推荐）

```go
package myext

import (
    "github.com/ncobase/ncore/extension/registry"
)

func init() {
    // 简单注册
    registry.Register(New())

    // 注册到特定组
    registry.RegisterToGroup(New(), "business")

    // 带弱依赖注册
    registry.RegisterToGroupWithWeakDeps(New(), "business", []string{"user", "tenant"})
}

func New() *MyExtension {
    return &MyExtension{}
}
```

### 3.2 手动注册

```go
func main() {
    mgr, err := manager.NewManager(config)
    if err != nil {
        panic(err)
    }

    // 手动注册扩展
    err = mgr.RegisterExtension(myext.New())
    if err != nil {
        panic(err)
    }

    // 初始化所有扩展
    if err := mgr.InitExtensions(); err != nil {
        panic(err)
    }
}
```

## 四、依赖管理

### 4.1 依赖类型

```go
// 强依赖 - 必须存在，否则初始化失败
type StrongDependency string = "strong"

// 弱依赖 - 可选存在，缺失时优雅降级
type WeakDependency string = "weak"

type DependencyEntry struct {
    Name string
    Type DependencyType
}
```

### 4.2 处理弱依赖

```go
func (m *MyExtension) PostInit() error {
    // 尝试获取可选服务
    if userService, err := m.manager.GetServiceByName("user"); err == nil {
        m.userService = userService
        m.log.Info("用户服务已连接")
    } else {
        m.log.Warn("用户服务不可用，部分功能受限")
    }

    // 订阅事件（如果事件系统可用）
    m.setupEventSubscriptions()

    return nil
}

func (m *MyExtension) setupEventSubscriptions() {
    // 订阅用户事件
    m.manager.SubscribeEvent("user.created", func(data any) {
        if m.userService != nil {
            m.handleUserCreated(data)
        }
    })
}
```

## 五、服务间通信

### 5.1 服务调用策略

```go
// 调用策略
const (
    LocalFirst  CallStrategy = iota // 本地优先
    RemoteFirst                     // 远程优先
    LocalOnly                       // 仅本地
    RemoteOnly                      // 仅远程
)

// 服务调用示例
func (s *MyService) CallUserService(ctx context.Context, userID string) (*User, error) {
    // 默认本地优先策略
    result, err := s.manager.CallService(ctx, "user", "GetUser", userID)
    if err != nil {
        return nil, err
    }

    return result.Response.(*User), nil
}

// 带选项调用
func (s *MyService) CallRemoteService(ctx context.Context) (*Response, error) {
    result, err := s.manager.CallServiceWithOptions(ctx, "external-service", "Process", nil,
        &types.CallOptions{
            Strategy: types.RemoteOnly,
            Timeout:  30 * time.Second,
        })

    return result.Response.(*Response), err
}
```

### 5.2 跨服务访问

```go
// 直接服务访问
userService, err := manager.GetServiceByName("user")
if err != nil {
    return err
}

// 跨服务字段访问
authService, err := manager.GetCrossService("auth", "TokenManager")
if err != nil {
    return err
}
```

### 5.3 熔断器保护

```go
func (s *MyService) CallExternalAPI() (*Response, error) {
    result, err := s.manager.ExecuteWithCircuitBreaker("external-api", func() (any, error) {
        // 调用外部 API
        return s.httpClient.Get("https://api.example.com/data")
    })

    if err != nil {
        return nil, err
    }

    return result.(*Response), nil
}
```

## 六、事件系统

### 6.1 事件目标

```go
const (
    EventTargetMemory EventTarget = 1 << iota // 内存事件总线
    EventTargetQueue                          // 消息队列
    EventTargetAll    = EventTargetMemory | EventTargetQueue // 所有目标
)
```

### 6.2 事件操作

```go
// 事件订阅
func (m *MyExtension) setupEvents() {
    // 订阅内存事件
    m.manager.SubscribeEvent("user.created", m.handleUserCreated, types.EventTargetMemory)

    // 订阅队列事件（分布式）
    m.manager.SubscribeEvent("payment.completed", m.handlePayment, types.EventTargetQueue)

    // 自动选择最佳传输
    m.manager.SubscribeEvent("system.alert", m.handleAlert)
}

// 事件发布
func (s *MyService) CreateUser(ctx context.Context, req *CreateUserReq) (*User, error) {
    user, err := s.userRepo.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    // 发布事件
    s.manager.PublishEvent("user.created", user)

    // 关键事件带重试
    s.manager.PublishEventWithRetry("user.profile.setup", user, 3)

    return user, nil
}
```

### 6.3 事件数据处理

```go
func (m *MyExtension) handleUserCreated(data any) {
    // 提取事件数据
    payload, err := types.ExtractEventPayload(data)
    if err != nil {
        m.log.Errorf("提取事件数据失败: %v", err)
        return
    }

    // 安全获取字段
    userID := types.SafeGet[string](payload, "user_id")
    userName := types.SafeGetWithDefault(payload, "name", "未知用户")

    // 处理业务逻辑
    m.processNewUser(userID, userName)
}
```

## 七、配置管理

### 7.1 扩展配置

```yaml
extension:
  path: "./plugins" # 插件目录
  mode: "file" # "file" 或 "c2hlbgo"（内置）
  includes: ["auth", "user"] # 包含特定插件
  excludes: ["debug"] # 排除插件

consul:
  address: "localhost:8500" # Consul 服务器
  scheme: "http"
  discovery:
    health_check: true # 启用健康检查
    check_interval: "10s" # 健康检查间隔
    timeout: "3s" # 健康检查超时

grpc:
  enabled: true # 启用 gRPC 支持
  host: "localhost"
  port: 9090
```

### 7.2 扩展专用配置

```go
// 配置结构
type MyExtensionConfig struct {
    Enabled    bool   `yaml:"enabled"`
    APIKey     string `yaml:"api_key"`
    BatchSize  int    `yaml:"batch_size"`
    Timeout    string `yaml:"timeout"`
}

// 在 Init 方法中读取配置
func (m *MyExtension) Init(conf *config.Config, mgr types.ManagerInterface) error {
    // 从配置中提取扩展专用配置
    var extConfig MyExtensionConfig
    if err := conf.UnmarshalKey("my_extension", &extConfig); err != nil {
        return fmt.Errorf("解析配置失败: %v", err)
    }

    m.config = &extConfig
    return nil
}
```

## 八、gRPC 集成

### 8.1 gRPC 服务定义

```protobuf
// proto/myservice.proto
syntax = "proto3";

package myservice;

service MyService {
  rpc GetData(GetDataRequest) returns (GetDataResponse);
  rpc ProcessData(ProcessDataRequest) returns (ProcessDataResponse);
}

message GetDataRequest {
  string id = 1;
}

message GetDataResponse {
  string data = 1;
  int64 timestamp = 2;
}
```

### 8.2 gRPC 服务实现

```go
// 实现 gRPC 扩展接口
func (m *MyExtension) RegisterGRPCServices(server *grpc.Server) {
    pb.RegisterMyServiceServer(server, m.grpcService)
}

// gRPC 服务实现
type grpcService struct {
    pb.UnimplementedMyServiceServer
    svc *MyService
}

func (g *grpcService) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
    data, err := g.svc.GetData(ctx, req.Id)
    if err != nil {
        return nil, err
    }

    return &pb.GetDataResponse{
        Data:      data.Content,
        Timestamp: data.CreatedAt.Unix(),
    }, nil
}
```

## 九、最佳实践

### 9.1 错误处理

```go
// 使用标准错误包装
import "fmt"

func (s *MyService) ProcessData(ctx context.Context, data *Data) error {
    if err := s.validate(data); err != nil {
        return fmt.Errorf("数据验证失败：%w", err)
    }

    if err := s.save(ctx, data); err != nil {
        return fmt.Errorf("保存数据失败：%w", err)
    }

    return nil
}
```

### 9.2 日志记录

```go
import "github.com/ncobase/ncore/logging/logger"

func (m *MyExtension) processRequest(ctx context.Context, req *Request) error {
    logger.Infof(ctx, "处理请求开始: %s", req.ID)

    start := time.Now()
    defer func() {
        logger.Infof(ctx, "处理请求完成: %s, 耗时: %v", req.ID, time.Since(start))
    }()

    if err := m.validate(req); err != nil {
        logger.Errorf(ctx, "请求验证失败: %v", err)
        return err
    }

    logger.Debugf(ctx, "请求参数：%+v", req)
    return m.execute(ctx, req)
}
```

### 9.3 优雅关闭

```go
func (m *MyExtension) PreCleanup() error {
    // 停止接收新请求
    m.stopAcceptingRequests()

    // 等待正在处理的请求完成
    m.waitForActiveRequests(30 * time.Second)

    return nil
}

func (m *MyExtension) Cleanup() error {
    // 关闭数据库连接
    if m.db != nil {
        m.db.Close()
    }

    // 关闭文件句柄
    if m.file != nil {
        m.file.Close()
    }

    // 取消上下文
    if m.cancel != nil {
        m.cancel()
    }

    return nil
}
```

## 十、管理 API

系统提供了完整的管理 API：

```bash
# 查看所有扩展
GET /api/exts

# 查看扩展状态
GET /api/exts/status

# 加载插件
POST /api/exts/load?name=plugin-name

# 卸载插件
POST /api/exts/unload?name=plugin-name

# 重载插件
POST /api/exts/reload?name=plugin-name

# 查看系统指标
GET /api/exts/metrics

# 查看特定指标
GET /api/exts/metrics/events
GET /api/exts/metrics/cache
GET /api/exts/metrics/extensions
```

## 十一、性能优化

### 11.1 服务发现缓存

```go
// 设置适当的缓存 TTL
manager.SetServiceCacheTTL(30 * time.Second)

// 定期清理缓存
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        manager.ClearServiceCache()
    }
}()
```

### 11.2 事件传输选择

```go
// 高频率事件使用内存传输
manager.PublishEvent("metrics.update", data, types.EventTargetMemory)

// 关键业务事件使用队列传输
manager.PublishEvent("order.created", order, types.EventTargetQueue)

// 系统级事件使用双重传输
manager.PublishEvent("system.shutdown", nil, types.EventTargetAll)
```

### 11.3 熔断器配置

```go
// 自定义熔断器设置
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "external-api",
    MaxRequests: 100,
    Interval:    5 * time.Second,
    Timeout:     3 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 3 && failureRatio >= 0.6
    },
})
```

## 十二、故障排除

### 12.1 常见问题

**循环依赖**

```text
错误: cyclic dependency detected in extensions: [module-a, module-b]
```

_解决方案_: 将其中一个依赖转换为弱依赖类型

**服务未找到**

```text
错误: extension 'user-service' not found
```

_解决方案_: 检查扩展注册和初始化顺序

**gRPC 连接失败**

```text
错误: failed to get gRPC connection for service-name
```

_解决方案_: 验证服务发现配置和网络连接

### 12.2 调试技巧

```go
// 启用调试日志
logger.SetLevel(logger.DebugLevel)

// 检查扩展状态
status := manager.GetStatus()
for name, state := range status {
    logger.Infof(nil, "扩展 %s 状态: %s", name, state)
}

// 检查依赖关系
metadata := manager.GetMetadata()
for name, meta := range metadata {
    logger.Infof(nil, "扩展 %s 依赖: %v", name, meta.Dependencies)
}
```

通过遵循本指南，您可以快速开发出结构清晰、功能完整的扩展模块。
