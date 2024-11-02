# 扩展（模块、插件）设计开发指南

Extensions Design Development Guide

> 本指南面向需要在该框架下进行模块或插件开发的开发人员，详细介绍了：
>
> - 标准的项目结构和模块组织方式
> - 模块开发的具体规范和最佳实践
> - 代码复用和测试相关的开发指导
>
> 通过遵循本指南，开发人员可以快速理解系统架构，开发出结构清晰、易于维护的模块代码。

## 一、项目目录结构

### 1. 顶层目录结构

```plaintext
.
├── bin/              # 编译产物目录
├── cmd/              # 命令行应用程序
├── core/             # 核心系统组件
├── docs/             # 文档
├── domain/           # 业务领域逻辑
├── front/            # 前端代码库
├── logs/             # 日志文件
├── pkg/              # 共享包和工具
├── plugin/           # 插件系统
├── proxy/            # API代理功能
└── setup/            # 系统安装和配置
```

### 2. 命令入口 (cmd)

```plaintext
cmd/
├── cli/              # 命令行工具
│   └── feature       # 特性管理
└── ncobase/          # 主程序
    ├── middleware    # 中间件
    └── service       # 服务配置
```

### 3. 核心模块 (core)

```plaintext
core/
├── access/           # 访问控制
├── auth/             # 认证授权
├── group/            # 组织管理
├── relation/         # 关系管理
├── system/           # 系统管理
├── tenant/           # 租户管理
└── user/             # 用户管理
```

### 4. 领域模块 (domain)

```plaintext
domain/
├── content/          # 内容管理
└── resource/         # 资源管理
```

### 5. 前端结构 (front)

```plaintext
front/
├── apps/             # 应用程序
│   └── console       # 控制台
├── infra/            # 前端基础设施
│   └── serverless    # 无服务架构
├── packages/         # 共享包
└── plugins/          # 前端插件
```

### 6. 基础设施 (infra)

```plaintext
infra/
├── config/           # 配置管理
├── database/         # 数据库配置
└── systemd/          # 系统服务
```

### 7. 公共包 (pkg)

```plaintext
pkg/
├── biz/              # 业务工具
├── cache/            # 缓存
├── config/           # 配置
├── data/             # 数据处理
├── ecode/            # 错误码
├── email/            # 邮件
├── feature/          # 特性开关
├── log/              # 日志
├── router/           # 路由
├── storage/          # 存储
└── validator/        # 验证器
```

### 8. 插件模块 (plugin)

```plaintext
plugin/
├── counter/          # 计数器插件
└── sample/           # 示例插件
```

> 完整版本请参考 [Overview](Overview.md)

## 二、模块标准结构

每个功能模块（core/、domain/、plugin/）下的标准结构：

```plaintext
[module_name]/
├── README.md           # 模块说明文档
├── [module].go         # 模块入口文件
├── services.go         # 服务集合声明
├── generate.go         # 代码生成入口
├── go.mod              # 模块依赖管理
├── go.sum              # 依赖版本锁定
├── config/             # 配置定义
│   └── options.go      # 配置选项
├── data/               # 数据访问层
│   ├── data.go         # 数据层入口
│   ├── ent/            # ORM实体
│   ├── repository/     # 数据仓储实现
│   └── schema/         # 数据库模式定义
├── handler/            # 请求处理层
│   ├── provider.go     # 处理器提供者
│   └── *.go            # 具体处理器
├── service/            # 业务逻辑层
│   ├── provider.go     # 服务提供者
│   ├── helper.go       # 辅助方法
│   └── *.go            # 具体服务实现
├── initialize/         # 初始化逻辑（可选）
│   ├── initialize.go   # 初始化入口
│   └── *.go            # 具体初始化任务
└── structs/            # 数据结构定义
    └── *.go            # 数据结构
```

### 示例：访问控制模块 (access) 结构

```plaintext
access/
├── access.go          # 模块入口
├── data/              # 数据访问层
│   ├── data.go
│   ├── ent/           # ORM
│   ├── repository/    # 仓储
│   └── schema/        # 数据模式
├── handler/           # 处理器
│   ├── casbin.go
│   ├── permission.go
│   ├── provider.go
│   ├── role.go
│   └── role_permission.go
├── service/           # 服务层
│   ├── casbin.go
│   ├── casbin_adapter.go
│   ├── helper.go
│   ├── permission.go
│   ├── provider.go
│   ├── role.go
│   └── role_permission.go
└── structs/           # 结构定义
    ├── casbin_rule.go
    ├── permission.go
    ├── role.go
    └── role_permission.go
```

### 示例：系统模块 (system) 结构

```plaintext
system/
├── README.md          # 说明文档
├── system.go          # 模块入口
├── services.go        # 服务集合
├── config/            # 配置
│   └── options.go
├── data/              # 数据层
│   ├── data.go
│   ├── ent/
│   ├── repository/
│   └── schema/
├── handler/           # 处理器
│   ├── dictionary.go
│   ├── instance.go
│   ├── menu.go
│   └── provider.go
├── initialize/        # 初始化
│   ├── initialize.go
│   ├── casbin.go
│   ├── group.go
│   └── user.go
├── service/           # 服务层
│   ├── dictionary.go
│   ├── helper.go
│   ├── menu.go
│   └── provider.go
└── structs/           # 结构定义
    ├── dictionary.go
    ├── menu.go
    └── options.go
```

## 三、API 领域划分

### 1. 身份与访问管理 (iam)

```plaintext
/iam/login              # 登录
/iam/logout             # 登出
/iam/users              # 用户管理
/iam/roles              # 角色管理
/iam/permissions        # 权限管理
/iam/tenants            # 租户管理
```

### 2. 系统管理 (sys)

```plaintext
/sys/menus              # 菜单管理
/sys/settings           # 系统设置
/sys/dictionaries       # 字典管理
```

> 更多 API 领域划分请参考 [业务领域参考](Business_Domain_Reference.md)

## 四、模块开发规范

### 1. 核心接口实现

所有模块必须实现 extension.Interface 接口：

```go
type Interface interface {
    Name() string                          // 模块名称
    PreInit() error                        // 初始化前的准备
    Init(*config.Config, *Manager) error   // 初始化
    PostInit() error                       // 初始化后的处理
    RegisterRoutes(*gin.RouterGroup)       // 注册路由
    GetHandlers() Handler                  // 获取处理器
    GetServices() Service                  // 获取服务
    PreCleanup() error                    // 清理前的处理
    Cleanup() error                       // 清理资源
    Status() string                       // 模块状态
    GetMetadata() Metadata                // 获取元数据
    Version() string                      // 版本信息
    Dependencies() []string               // 依赖列表
}
```

### 2. 分层设计规范

#### 2.1 Repository 层

```go
// 仓储基础接口
type BaseRepository interface {
    WithContext(ctx context.Context) BaseRepository
    WithTenant(tenantID string) BaseRepository
    WithTransaction(tx *ent.Tx) BaseRepository
}

// 实体仓储接口
type EntityRepository interface {
    BaseRepository
    Create(ctx context.Context, entity *Entity) error
    Update(ctx context.Context, entity *Entity) error
    Delete(ctx context.Context, id string) error
    Get(ctx context.Context, id string) (*Entity, error)
    List(ctx context.Context, params *ListParams) ([]*Entity, error)
}

// 仓储实现示例
type repository struct {
    data *data.Data
    log  *log.Logger
}

func (r *repository) WithContext(ctx context.Context) BaseRepository {
    return &repository{data: r.data.WithContext(ctx)}
}
```

#### 2.2 Service 层

```go
// 服务基础接口
type BaseService interface {
    WithContext(ctx context.Context) BaseService
    WithTenant(tenantID string) BaseService
}

// 实体服务接口
type EntityService interface {
    BaseService
    Create(ctx context.Context, req *CreateReq) (*Entity, error)
    Update(ctx context.Context, req *UpdateReq) (*Entity, error)
    Delete(ctx context.Context, id string) error
    Get(ctx context.Context, id string) (*Entity, error)
    List(ctx context.Context, params *ListParams) (*ListResult, error)
}

// 服务实现示例
type service struct {
    repo     EntityRepository
    eventBus *EventBus
    cache    Cache
}

func (s *service) WithContext(ctx context.Context) BaseService {
    return &service{
        repo: s.repo.WithContext(ctx),
    }
}
```

#### 2.3 Handler 层

```go
// 处理器基础接口
type BaseHandler interface {
    Create(c *gin.Context)
    Update(c *gin.Context)
    Delete(c *gin.Context)
    Get(c *gin.Context)
    List(c *gin.Context)
}

// 处理器实现示例
type handler struct {
    service EntityService
    log     *log.Logger
}

func (h *handler) Create(c *gin.Context) {
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, err)
        return
    }

    ctx := c.Request.Context()
    result, err := h.service.
        WithContext(ctx).
        WithTenant(h.getTenantID(c)).
        Create(ctx, &req)

    if err != nil {
        c.JSON(500, err)
        return
    }

    c.JSON(200, result)
}
```

## 五、模块复用策略

### 1. 依赖注入与服务发现

```go
// 获取依赖服务
func (m *Module) getDependencyService() error {
    // 获取认证服务
    authService, err := m.getService("auth")
    if err != nil {
        return fmt.Errorf("failed to get auth service: %v", err)
    }

    // 获取用户服务
    userService, err := m.getService("user")
    if err != nil {
        return fmt.Errorf("failed to get user service: %v", err)
    }

    m.auth = authService
    m.user = userService
    return nil
}
```

### 2. 事件驱动通信

```go
// 事件订阅
func (m *Module) subscribeEvents() {
    // 订阅用户创建事件
    m.em.SubscribeEvent("user.created", func(data any) {
        user := data.(*User)
        // 处理用户创建后的业务逻辑
    })

    // 订阅租户变更事件
    m.em.SubscribeEvent("tenant.changed", func(data any) {
        tenant := data.(*Tenant)
        // 处理租户变更后的业务逻辑
    })
}

// 事件发布
func (s *service) Create(ctx context.Context, req *CreateReq) (*Entity, error) {
    // 创建实体
    entity, err := s.repo.Create(ctx, req.toEntity())
    if err != nil {
        return nil, err
    }

    // 发布创建事件
    s.eventBus.Publish("entity.created", entity)

    return entity, nil
}
```

### 3. 中间件复用

```go
// 通用中间件
func AuthMiddleware(authSvc AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        user, err := authSvc.ValidateToken(token)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
        c.Set("user", user)
        c.Next()
    }
}

func TenantMiddleware(tenantSvc TenantService) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetHeader("X-Tenant-ID")
        tenant, err := tenantSvc.ValidateTenant(tenantID)
        if err != nil {
            c.AbortWithStatus(403)
            return
        }
        c.Set("tenant", tenant)
        c.Next()
    }
}

// 使用中间件
func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    r.Use(
        AuthMiddleware(m.authSvc),
        TenantMiddleware(m.tenantSvc),
    )
    // 注册路由
}
```

## 六、测试规范

### 1. 单元测试

```go
func TestService_Create(t *testing.T) {
    // 准备测试用例
    tests := []struct {
        name    string
        req     *CreateReq
        mock    func(*mockRepo)
        want    *Entity
        wantErr bool
    }{
        {
            name: "success",
            req: &CreateReq{Name: "test"},
            mock: func(m *mockRepo) {
                m.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    Return(&Entity{ID: "1", Name: "test"}, nil)
            },
            want: &Entity{ID: "1", Name: "test"},
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 初始化 mock
            ctrl := gomock.NewController(t)
            mockRepo := NewMockRepository(ctrl)
            tt.mock(mockRepo)

            // 创建服务实例
            svc := NewService(mockRepo)

            // 执行测试
            got, err := svc.Create(context.Background(), tt.req)

            // 验证结果
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 2. 集成测试

```go
func TestAPI_Integration(t *testing.T) {
    // 初始化测试环境
    app := newTestApp(t)
    defer app.Cleanup()

    // 准备测试数据
    tenant := createTestTenant(t, app)
    user := createTestUser(t, app)

    // 测试完整流程
    t.Run("full workflow", func(t *testing.T) {
        // 创建实体
        entity := createEntity(t, app, &CreateReq{
            Name:     "test",
            TenantID: tenant.ID,
        })

        // 验证创建结果
        assert.NotEmpty(t, entity.ID)
        assert.Equal(t, "test", entity.Name)

        // 更新实体
        updated := updateEntity(t, app, entity.ID, &UpdateReq{
            Name: "updated",
        })
        assert.Equal(t, "updated", updated.Name)

        // 删除实体
        deleteEntity(t, app, entity.ID)

        // 验证删除结果
        assertNotFound(t, app, entity.ID)
    })
}

// 测试助手函数
func createEntity(t *testing.T, app *TestApp, req *CreateReq) *Entity {
    resp := app.Post("/api/entities", req)
    require.Equal(t, 200, resp.Code)

    var entity Entity
    require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &entity))
    return &entity
}
```
