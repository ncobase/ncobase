package repository

import (
	"context"
	"ncobase/common/data/cache"
	"ncobase/common/data/meili"
	"ncobase/common/log"
	"ncobase/common/validator"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	templateEnt "ncobase/core/workflow/data/ent/template"
	"ncobase/core/workflow/structs"

	"github.com/redis/go-redis/v9"
)

type TemplateRepositoryInterface interface {
	Create(context.Context, *structs.TemplateBody) (*ent.Template, error)
	Get(context.Context, *structs.FindTemplateParams) (*ent.Template, error)
	Update(context.Context, *structs.UpdateTemplateBody) (*ent.Template, error)
	Delete(context.Context, *structs.FindTemplateParams) error
	List(context.Context, *structs.ListTemplateParams) ([]*ent.Template, error)
	CountX(context.Context, *structs.ListTemplateParams) int

	GetLatestVersion(context.Context, string) (*ent.Template, error)
	MarkAsLatest(context.Context, string) error
	DisableTemplate(context.Context, string) error
}

type templateRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Template]
}

func NewTemplateRepository(d *data.Data) TemplateRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &templateRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Template](rc, "workflow_template", false),
	}
}

// Create creates a new template
func (r *templateRepository) Create(ctx context.Context, body *structs.TemplateBody) (*ent.Template, error) {
	builder := r.ec.Template.Create()

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Code != "" {
		builder.SetCode(body.Code)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}
	if body.Type != "" {
		builder.SetType(body.Type)
	}
	if body.Version != "" {
		builder.SetVersion(body.Version)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.ModuleCode != "" {
		builder.SetModuleCode(body.ModuleCode)
	}
	if body.FormCode != "" {
		builder.SetFormCode(body.FormCode)
	}
	if body.TemplateKey != "" {
		builder.SetTemplateKey(body.TemplateKey)
	}
	if body.Category != "" {
		builder.SetCategory(body.Category)
	}

	// Set JSON fields
	if body.NodeConfig != nil {
		builder.SetNodeConfig(body.NodeConfig)
	}
	if body.NodeRules != nil {
		builder.SetNodeRules(body.NodeRules)
	}
	if body.NodeEvents != nil {
		builder.SetNodeEvents(body.NodeEvents)
	}
	if body.ProcessRules != nil {
		builder.SetProcessRules(body.ProcessRules)
	}
	if body.FormConfig != nil {
		builder.SetFormConfig(body.FormConfig)
	}
	if body.FormPerms != nil {
		builder.SetFormPermissions(body.FormPerms)
	}
	if body.RoleConfigs != nil {
		builder.SetRoleConfigs(body.RoleConfigs)
	}

	// Set boolean fields
	builder.SetIsDraftEnabled(body.IsDraftEnabled)
	builder.SetIsAutoStart(body.IsAutoStart)
	builder.SetStrictMode(body.StrictMode)
	builder.SetAllowCancel(body.AllowCancel)
	builder.SetAllowUrge(body.AllowUrge)
	builder.SetAllowDelegate(body.AllowDelegate)
	builder.SetAllowTransfer(body.AllowTransfer)
	builder.SetIsLatest(body.IsLatest)
	builder.SetDisabled(body.Disabled)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "templateRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("templates", row); err != nil {
		log.Errorf(ctx, "templateRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific template
func (r *templateRepository) Get(ctx context.Context, params *structs.FindTemplateParams) (*ent.Template, error) {
	builder := r.ec.Template.Query()

	if params.Code != "" {
		builder.Where(templateEnt.CodeEQ(params.Code))
	}
	if params.ModuleCode != "" {
		builder.Where(templateEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(templateEnt.FormCodeEQ(params.FormCode))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(templateEnt.StatusEQ(params.Status))
	}
	if params.Type != "" {
		builder.Where(templateEnt.TypeEQ(params.Type))
	}
	if params.Version != "" {
		builder.Where(templateEnt.VersionEQ(params.Version))
	}
	if params.Category != "" {
		builder.Where(templateEnt.CategoryEQ(params.Category))
	}
	if params.IsLatest != nil {
		builder.Where(templateEnt.IsLatestEQ(*params.IsLatest))
	}
	if params.Disabled != nil {
		builder.Where(templateEnt.DisabledEQ(*params.Disabled))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing template
func (r *templateRepository) Update(ctx context.Context, body *structs.UpdateTemplateBody) (*ent.Template, error) {
	builder := r.ec.Template.UpdateOneID(body.ID)

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.Version != "" {
		builder.SetVersion(body.Version)
	}

	// Update JSON fields
	if body.NodeConfig != nil {
		builder.SetNodeConfig(body.NodeConfig)
	}
	if body.ProcessRules != nil {
		builder.SetProcessRules(body.ProcessRules)
	}
	if body.FormConfig != nil {
		builder.SetFormConfig(body.FormConfig)
	}

	// Update boolean fields
	builder.SetIsDraftEnabled(body.IsDraftEnabled)
	builder.SetIsAutoStart(body.IsAutoStart)
	builder.SetStrictMode(body.StrictMode)
	builder.SetIsLatest(body.IsLatest)
	builder.SetDisabled(body.Disabled)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// Delete deletes a template
func (r *templateRepository) Delete(ctx context.Context, params *structs.FindTemplateParams) error {
	builder := r.ec.Template.Delete()

	if params.Code != "" {
		builder.Where(templateEnt.CodeEQ(params.Code))
	}

	_, err := builder.Exec(ctx)
	return err
}

// List returns a list of templates
func (r *templateRepository) List(ctx context.Context, params *structs.ListTemplateParams) ([]*ent.Template, error) {
	builder := r.ec.Template.Query()

	if params.ModuleCode != "" {
		builder.Where(templateEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(templateEnt.FormCodeEQ(params.FormCode))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(templateEnt.StatusEQ(params.Status))
	}
	if params.Type != "" {
		builder.Where(templateEnt.TypeEQ(params.Type))
	}
	if params.Category != "" {
		builder.Where(templateEnt.CategoryEQ(params.Category))
	}
	if params.IsLatest != nil {
		builder.Where(templateEnt.IsLatestEQ(*params.IsLatest))
	}

	// Add sorting
	switch params.SortBy {
	case structs.SortByName:
		builder.Order(ent.Asc(templateEnt.FieldName))
	default:
		builder.Order(ent.Desc(templateEnt.FieldCreatedAt))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CountX returns the total count of templates
func (r *templateRepository) CountX(ctx context.Context, params *structs.ListTemplateParams) int {
	builder := r.ec.Template.Query()

	if params.ModuleCode != "" {
		builder.Where(templateEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(templateEnt.FormCodeEQ(params.FormCode))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(templateEnt.StatusEQ(params.Status))
	}
	if params.Type != "" {
		builder.Where(templateEnt.TypeEQ(params.Type))
	}

	return builder.CountX(ctx)
}

// GetLatestVersion returns the latest version of a template
func (r *templateRepository) GetLatestVersion(ctx context.Context, code string) (*ent.Template, error) {
	return r.ec.Template.Query().
		Where(templateEnt.CodeEQ(code)).
		Where(templateEnt.IsLatestEQ(true)).
		First(ctx)
}

// MarkAsLatest marks a template version as the latest
func (r *templateRepository) MarkAsLatest(ctx context.Context, templateID string) error {
	// First, unmark all versions of this template
	template, err := r.ec.Template.Get(ctx, templateID)
	if err != nil {
		return err
	}

	_, err = r.ec.Template.Update().
		Where(templateEnt.CodeEQ(template.Code)).
		SetIsLatest(false).
		Save(ctx)
	if err != nil {
		return err
	}

	// Then mark this version as latest
	return r.ec.Template.UpdateOneID(templateID).
		SetIsLatest(true).
		Exec(ctx)
}

// DisableTemplate disables a template
func (r *templateRepository) DisableTemplate(ctx context.Context, templateID string) error {
	return r.ec.Template.UpdateOneID(templateID).
		SetDisabled(true).
		Exec(ctx)
}
