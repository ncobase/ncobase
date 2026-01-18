# Extension Guide

Guide for developing modules and plugins in ncobase using the ncore extension system.

## Extension Types

| Type     | Location  | Registration               | Use Case                    |
| -------- | --------- | -------------------------- | --------------------------- |
| Module   | `core/`   | `registry.RegisterToGroup` | Core system functionality   |
| Business | `biz/`    | `registry.RegisterToGroup` | Business logic modules      |
| Plugin   | `plugin/` | `plugin.RegisterPlugin`    | Optional/pluggable features |

## Core Interface

All extensions must implement `ext.Interface`:

```go
type Interface interface {
    Name() string
    Version() string
    Init(*config.Config, ManagerInterface) error
    GetMetadata() Metadata
    GetHandlers() Handler
    GetServices() Service
    Dependencies() []string
    OptionalMethods
}
```

## Module Template

```go
package mymodule

import (
    "fmt"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/ncobase/ncore/config"
    exr "github.com/ncobase/ncore/extension/registry"
    ext "github.com/ncobase/ncore/extension/types"
)

var (
    name    = "mymodule"
    desc    = "My module description"
    version = "1.0.0"
    typeStr = "module"
    group   = "sys"
)

type Module struct {
    ext.OptionalImpl

    initialized bool
    mu          sync.RWMutex
    em          ext.ManagerInterface
    conf        *config.Config
    cleanup     func(name ...string)

    h *handler.Handler
    s *service.Service
    d *data.Data
}

func init() {
    exr.RegisterToGroupWithWeakDeps(New(), group, []string{})
}

func New() ext.Interface {
    return &Module{}
}

func (m *Module) Name() string    { return name }
func (m *Module) Version() string { return version }
func (m *Module) Type() string    { return typeStr }
func (m *Module) Group() string   { return group }

func (m *Module) Dependencies() []string {
    return []string{} // Strong dependencies
}

func (m *Module) Init(conf *config.Config, em ext.ManagerInterface) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.initialized {
        return fmt.Errorf("%s module already initialized", name)
    }

    // Initialize data layer
    var err error
    m.d, m.cleanup, err = data.New(conf.Data, conf.Environment)
    if err != nil {
        return err
    }

    m.em = em
    m.conf = conf
    m.initialized = true
    return nil
}

func (m *Module) PostInit() error {
    m.s = service.New(m.conf, m.d)
    m.h = handler.New(m.s)
    return nil
}

func (m *Module) RegisterRoutes(r *gin.RouterGroup) {
    api := r.Group("/" + m.Group())
    {
        api.GET("/items", m.h.List)
        api.POST("/items", m.h.Create)
        api.GET("/items/:id", m.h.Get)
        api.PUT("/items/:id", m.h.Update)
        api.DELETE("/items/:id", m.h.Delete)
    }
}

func (m *Module) GetHandlers() ext.Handler { return m.h }
func (m *Module) GetServices() ext.Service { return m.s }

func (m *Module) GetMetadata() ext.Metadata {
    return ext.Metadata{
        Name:         name,
        Version:      version,
        Dependencies: m.Dependencies(),
        Description:  desc,
        Type:         typeStr,
        Group:        group,
    }
}

func (m *Module) Cleanup() error {
    if m.cleanup != nil {
        m.cleanup(m.Name())
    }
    return nil
}
```

## Plugin Template

```go
package myplugin

import (
    "github.com/ncobase/ncore/config"
    extp "github.com/ncobase/ncore/extension/plugin"
    ext "github.com/ncobase/ncore/extension/types"
)

var (
    name    = "myplugin"
    desc    = "My plugin description"
    version = "1.0.0"
    typeStr = "plugin"
    group   = "plug"
)

type Plugin struct {
    ext.OptionalImpl
    // ... fields
}

func init() {
    extp.RegisterPlugin(New(), ext.Metadata{
        Name:        name,
        Version:     version,
        Description: desc,
        Type:        typeStr,
        Group:       group,
    })
}

func New() ext.Interface {
    return &Plugin{}
}

// Implement Interface methods...
```

## Directory Structure

```text
mymodule/
├── mymodule.go          # Extension entry point
├── data/
│   ├── data.go          # Data layer initialization
│   ├── ent/             # Ent schema and generated code
│   │   └── schema/
│   └── repository/      # Repository implementations
├── handler/
│   ├── handler.go       # Handler initialization
│   └── item.go          # HTTP handlers
├── service/
│   ├── service.go       # Service initialization
│   └── item.go          # Business logic
├── structs/
│   └── item.go          # Request/response structs
└── event/
    └── registrar.go     # Event handlers
```

## Data Layer

```go
package data

import (
    "github.com/ncobase/ncore/config"
    "github.com/ncobase/ncore/data"
)

type Data struct {
    *data.Data
    EC     *ent.Client
    ECRead *ent.Client
}

func New(conf *config.Data, env ...string) (*Data, func(), error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, nil, err
    }

    masterDB := d.GetMasterDB()
    entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, env...)
    if err != nil {
        return nil, cleanup, err
    }

    var entClientRead *ent.Client
    if readDB, err := d.GetSlaveDB(); err == nil && readDB != masterDB {
        entClientRead, _ = newEntClient(readDB, conf.Database.Master, false, env...)
    }
    if entClientRead == nil {
        entClientRead = entClient
    }

    return &Data{Data: d, EC: entClient, ECRead: entClientRead}, cleanup, nil
}

func (d *Data) GetMasterEntClient() *ent.Client { return d.EC }
func (d *Data) GetSlaveEntClient() *ent.Client  { return d.ECRead }
```

## Service Discovery

```go
func (m *Module) NeedServiceDiscovery() bool {
    return true
}

func (m *Module) GetServiceInfo() *ext.ServiceInfo {
    if !m.NeedServiceDiscovery() {
        return nil
    }
    return &ext.ServiceInfo{
        Address: m.discovery.address,
        Tags:    []string{m.Group(), m.Type()},
        Meta: map[string]string{
            "name":    m.Name(),
            "version": m.Version(),
        },
    }
}
```

## Event Handling

```go
// event/registrar.go
package event

import ext "github.com/ncobase/ncore/extension/types"

type Registrar struct {
    em ext.ManagerInterface
}

func NewRegistrar(em ext.ManagerInterface) *Registrar {
    return &Registrar{em: em}
}

func (r *Registrar) RegisterHandlers(provider *EventProvider) {
    r.em.SubscribeEvent("user.created", provider.OnUserCreated)
    r.em.SubscribeEvent("user.deleted", provider.OnUserDeleted)
}

// handler/event.go
type EventProvider struct {
    svc *service.Service
}

func NewEventProvider(svc *service.Service) *EventProvider {
    return &EventProvider{svc: svc}
}

func (p *EventProvider) OnUserCreated(data any) {
    // Handle user created event
}
```

## Cross-Service Communication

```go
// Get service by name
userService, err := m.em.GetServiceByName("user")

// Call service method
result, err := m.em.CallService(ctx, "user", "GetUser", userID)

// Get cross-service field
authService, err := m.em.GetCrossService("auth", "TokenManager")
```

## Configuration

```yaml
# config.yaml
extension:
  path: "./plugins"
  mode: "file" # "file" or built-in
  includes: [] # Include specific extensions
  excludes: [] # Exclude specific extensions

consul:
  address: "localhost:8500"
  discovery:
    health_check: true
    check_interval: "10s"
```

## Lifecycle Methods

| Method       | When Called               | Purpose                     |
| ------------ | ------------------------- | --------------------------- |
| `PreInit`    | Before Init               | Pre-initialization setup    |
| `Init`       | During startup            | Initialize resources        |
| `PostInit`   | After all extensions init | Cross-extension setup       |
| `PreCleanup` | Before cleanup            | Stop accepting new requests |
| `Cleanup`    | During shutdown           | Release resources           |

## Registration Methods

```go
// Module registration (core/, biz/)
exr.Register(New())
exr.RegisterToGroup(New(), "sys")
exr.RegisterToGroupWithWeakDeps(New(), "sys", []string{"user"})

// Plugin registration (plugin/)
extp.RegisterPlugin(New(), metadata)
```

## Dependency Types

```go
// Strong dependency - required, fails if missing
func (m *Module) Dependencies() []string {
    return []string{"auth", "user"}
}

// Weak dependency - optional, graceful degradation
exr.RegisterToGroupWithWeakDeps(New(), group, []string{"analytics"})

// Check weak dependency in PostInit
func (m *Module) PostInit() error {
    if svc, err := m.em.GetServiceByName("analytics"); err == nil {
        m.analytics = svc
    }
    return nil
}
```
