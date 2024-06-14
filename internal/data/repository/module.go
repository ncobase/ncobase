package repo

import (
	"context"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	moduleEnt "stocms/internal/data/ent/module"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/meili"
	"stocms/pkg/types"
	"stocms/pkg/validator"
	"time"

	"github.com/redis/go-redis/v9"
)

// Module represents the module repository interface.
type Module interface {
	Create(ctx context.Context, body *structs.CreateModuleBody) (*ent.Module, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Module, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Module, error)
	List(ctx context.Context, params *structs.ListModuleParams) ([]*ent.Module, error)
	Delete(ctx context.Context, slug string) error
}

// moduleRepo implements the Module interface.
type moduleRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Module]
}

// NewModule creates a new module repository.
func NewModule(d *data.Data) Module {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &moduleRepo{ec, rc, ms, cache.NewCache[ent.Module](rc, cache.Key("sc_module"), true)}
}

// Create creates a new module.
func (r *moduleRepo) Create(ctx context.Context, body *structs.CreateModuleBody) (*ent.Module, error) {
	// create builder.
	builder := r.ec.Module.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableContent(&body.Content)
	builder.SetNillableThumbnail(&body.Thumbnail)
	builder.SetNillableTemp(body.Temp)
	builder.SetNillableMarkdown(body.Markdown)
	builder.SetNillablePrivate(body.Private)
	builder.SetNillableStatus(body.Status)
	builder.SetNillableReleased(body.Released)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "moduleRepo.Create error: %v\n", err)
		return nil, err
	}

	// Create the module in Meilisearch index
	if err = r.ms.IndexDocuments("modules", row); err != nil {
		log.Errorf(nil, "moduleRepo.Create error creating Meilisearch index: %v\n", err)
	}

	return row, nil
}

// GetBySlug gets a module by slug.
func (r *moduleRepo) GetBySlug(ctx context.Context, slug string) (*ent.Module, error) {
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.FindModule(ctx, &structs.FindModule{Slug: slug})
	if err != nil {
		log.Errorf(nil, "moduleRepo.GetBySlug error: %v\n", err)
		return nil, err
	}

	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "moduleRepo.GetBySlug cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a module (full or partial).
func (r *moduleRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Module, error) {
	module, err := r.FindModule(ctx, &structs.FindModule{Slug: slug})
	if err != nil {
		return nil, err
	}

	builder := module.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "content":
			builder.SetNillableContent(types.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(types.ToPointer(value.(string)))
		case "temp":
			builder.SetTemp(value.(bool))
		case "markdown":
			builder.SetMarkdown(value.(bool))
		case "private":
			builder.SetPrivate(value.(bool))
		case "status":
			builder.SetStatus(value.(int))
		case "released":
			builder.SetNillableReleased(types.ToPointer(value.(time.Time)))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "moduleRepo.Update error: %v\n", err)
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s", module.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, module.Slug)
	if err != nil {
		log.Errorf(nil, "moduleRepo.Update cache error: %v\n", err)
	}

	if err = r.ms.DeleteDocuments("modules", slug); err != nil {
		log.Errorf(nil, "moduleRepo.Update error deleting Meilisearch index: %v\n", err)
	}

	if err = r.ms.IndexDocuments("modules", row); err != nil {
		log.Errorf(nil, "moduleRepo.Update error updating Meilisearch index: %v\n", err)
	}

	return row, nil
}

// List gets a list of modules.
func (r *moduleRepo) List(ctx context.Context, p *structs.ListModuleParams) ([]*ent.Module, error) {
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	builder.Limit(int(p.Limit))

	builder.Order(ent.Desc(moduleEnt.FieldCreatedAt))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(nil, "moduleRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a module.
func (r *moduleRepo) Delete(ctx context.Context, slug string) error {
	module, err := r.FindModule(ctx, &structs.FindModule{Slug: slug})
	if err != nil {
		return err
	}

	builder := r.ec.Module.Delete()

	if _, err = builder.Where(moduleEnt.IDEQ(module.ID)).Exec(ctx); err != nil {
		log.Errorf(nil, "moduleRepo.Delete error: %v\n", err)
		return err
	}

	cacheKey := fmt.Sprintf("%s", module.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("module:slug:%s", module.Slug))
	if err != nil {
		log.Errorf(nil, "moduleRepo.Delete cache error: %v\n", err)
	}

	if err = r.ms.DeleteDocuments("modules", module.ID); err != nil {
		log.Errorf(nil, "moduleRepo.Delete index error: %v\n", err)
	}

	return nil
}

// FindModule finds a module.
func (r *moduleRepo) FindModule(ctx context.Context, p *structs.FindModule) (*ent.Module, error) {
	builder := r.ec.Module.Query()
	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(moduleEnt.IDEQ(p.ID))
	}

	if validator.IsNotEmpty(p.Slug) {
		builder = builder.Where(moduleEnt.Or(
			moduleEnt.ID(p.Slug),
			moduleEnt.SlugEQ(p.Slug),
		))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder creates list builder.
func (r *moduleRepo) ListBuilder(ctx context.Context, p *structs.ListModuleParams) (*ent.ModuleQuery, error) {
	var next *ent.Module
	if validator.IsNotEmpty(p.Cursor) {
		row, err := r.FindModule(ctx, &structs.FindModule{ID: p.Cursor})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}
	builder := r.ec.Module.Query()

	if next != nil {
		builder.Where(moduleEnt.CreatedAtLT(next.CreatedAt))
	}

	return builder, nil
}

// CountX gets a count of modules.
func (r *moduleRepo) CountX(ctx context.Context, p *structs.ListModuleParams) int {
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// Count gets a count of modules.
func (r *moduleRepo) Count(ctx context.Context, p *structs.ListModuleParams) (int, error) {
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}
