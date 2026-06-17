package ai

import (
	"fmt"
	systemWrapper "ncobase/core/system/wrapper"
	"ncobase/plugin/ai/data"
	"ncobase/plugin/ai/handler"
	"ncobase/plugin/ai/router"
	"ncobase/plugin/ai/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/config"
	extp "github.com/ncobase/ncore/extension/plugin"
	ext "github.com/ncobase/ncore/extension/types"
)

var (
	name         = "ai"
	desc         = "AI gateway plugin backed by deebus providers"
	version      = "1.0.0"
	dependencies []string
	typeStr      = "plugin"
	group        = "ai"
)

type Plugin struct {
	ext.OptionalImpl

	initialized bool
	mu          sync.RWMutex
	em          ext.ManagerInterface
	cleanup     func(name ...string)

	d *data.Data
	s *service.Service
	h *handler.Handler
	r *router.Router

	discovery
}

type discovery struct {
	address string
	tags    []string
	meta    map[string]string
}

func init() {
	extp.RegisterPlugin(New(), ext.Metadata{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
		Description:  desc,
		Type:         typeStr,
		Group:        group,
	})
}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return name
}

func (p *Plugin) PreInit() error {
	return nil
}

func (p *Plugin) Init(conf *config.Config, em ext.ManagerInterface) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("AI plugin already initialized")
	}

	p.d, p.cleanup, err = data.New(conf.Data, conf.Environment)
	if err != nil {
		return err
	}

	if conf.Consul != nil {
		p.discovery.address = conf.Consul.Address
		p.discovery.tags = conf.Consul.Discovery.DefaultTags
		p.discovery.meta = conf.Consul.Discovery.DefaultMeta
	}

	p.em = em
	p.initialized = true
	return nil
}

func (p *Plugin) PostInit() error {
	optionWrapper := systemWrapper.NewOptionServiceWrapper(p.em)
	configProvider := service.NewSystemOptionConfigProvider(optionWrapper)
	p.s = service.New(p.d, configProvider)
	p.h = handler.New(p.s)
	p.r = router.New(p.h)
	return nil
}

func (p *Plugin) RegisterRoutes(r *gin.RouterGroup) {
	p.r.Register(r, p.Group())
}

func (p *Plugin) GetHandlers() ext.Handler {
	return p.h
}

func (p *Plugin) GetServices() ext.Service {
	return p.s
}

func (p *Plugin) Cleanup() error {
	if p.cleanup != nil {
		p.cleanup(p.Name())
	}
	return nil
}

func (p *Plugin) GetMetadata() ext.Metadata {
	return ext.Metadata{
		Name:         p.Name(),
		Version:      p.Version(),
		Dependencies: p.Dependencies(),
		Description:  p.Description(),
		Type:         p.Type(),
		Group:        p.Group(),
	}
}

func (p *Plugin) Version() string {
	return version
}

func (p *Plugin) Dependencies() []string {
	return dependencies
}

func (p *Plugin) Description() string {
	return desc
}

func (p *Plugin) Type() string {
	return typeStr
}

func (p *Plugin) Group() string {
	return group
}

func (p *Plugin) GetServiceInfo() *ext.ServiceInfo {
	if !p.NeedServiceDiscovery() {
		return nil
	}

	metadata := p.GetMetadata()
	tags := append(p.discovery.tags, metadata.Group, metadata.Type)
	meta := make(map[string]string)
	for k, v := range p.discovery.meta {
		meta[k] = v
	}
	meta["name"] = metadata.Name
	meta["version"] = metadata.Version
	meta["group"] = metadata.Group
	meta["type"] = metadata.Type
	meta["description"] = metadata.Description

	return &ext.ServiceInfo{
		Address: p.discovery.address,
		Tags:    tags,
		Meta:    meta,
	}
}
