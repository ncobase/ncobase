package generator

import (
	"fmt"
	"ncobase/cmd/cli/generator/templates"
	"ncobase/cmd/cli/utils"
	"path/filepath"
)

// Options defines generation options
type Options struct {
	Name     string
	Type     string // core / business / plugin
	UseEnt   bool
	WithTest bool
	Group    string
}

// Generate generates code based on options
func Generate(opts *Options) error {
	if !utils.ValidateName(opts.Name) {
		return fmt.Errorf("invalid name: %s", opts.Name)
	}

	var basePath string
	var moduleType string
	var mainTemplate func(string) string

	switch opts.Type {
	case "core":
		basePath = filepath.Join("core", opts.Name)
		moduleType = "core"
		mainTemplate = templates.CoreTemplate
	case "module":
		basePath = filepath.Join("domain", opts.Name)
		moduleType = "domain"
		mainTemplate = templates.BusinessTemplate
	case "plugin":
		basePath = filepath.Join("plugin", opts.Name)
		moduleType = "plugin"
		mainTemplate = templates.PluginTemplate
	default:
		return fmt.Errorf("unknown type: %s", opts.Type)
	}

	// Check if component already exists
	if exists, err := utils.PathExists(basePath); err != nil {
		return fmt.Errorf("error checking existence: %v", err)
	} else if exists {
		return fmt.Errorf("'%s' already exists in %s", opts.Name, filepath.Dir(basePath))
	}

	// Prepare template data
	data := &templates.Data{
		Name:        opts.Name,
		Type:        opts.Type,
		Group:       opts.Group,
		UseEnt:      opts.UseEnt,
		WithTest:    opts.WithTest,
		ModuleType:  moduleType,
		PackagePath: fmt.Sprintf("ncobase/%s/%s", moduleType, opts.Name),
	}

	return createStructure(basePath, data, mainTemplate)
}

func createStructure(basePath string, data *templates.Data, mainTemplate func(string) string) error {
	// Create base directory
	if err := utils.EnsureDir(basePath); err != nil {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	// Create directory structure
	directories := []string{
		"data",
		"data/repository",
		"data/schema",
		"handler",
		"service",
		"structs",
	}

	if data.WithTest {
		directories = append(directories, "tests")
	}

	for _, dir := range directories {
		if err := utils.EnsureDir(filepath.Join(basePath, dir)); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	selectDataTemplate := func(data templates.Data) string {
		if data.UseEnt {
			return templates.DataTemplateWithEnt(data.Name, data.ModuleType)
		}
		return templates.DataTemplate(data.Name, data.ModuleType)
	}

	files := map[string]string{
		fmt.Sprintf("%s.go", data.Name): mainTemplate(data.Name),
		// "go.mod":                        templates.ModuleTemplate(data.Name, data.ModuleType),
		// "generate.go":                   templates.GeneraterTemplate(data.Name, data.ModuleType),
		"data/data.go":                  selectDataTemplate(*data),
		"data/repository/repository.go": templates.RepositoryTemplate(data.Name, data.ModuleType),
		"data/schema/schema.go":         templates.SchemaTemplate(),
		"handler/handler.go":            templates.HandlerTemplate(data.Name, data.ModuleType),
		"service/service.go":            templates.ServiceTemplate(data.Name, data.ModuleType),
		"structs/structs.go":            templates.StructsTemplate(),
	}

	// Add ent files if required
	if data.UseEnt {
		files["generate.go"] = templates.GeneraterTemplate(data.Name, data.ModuleType)
	}

	// Add test files if required
	if data.WithTest {
		files["tests/module_test.go"] = templates.ModuleTestTemplate(data.Name, data.ModuleType)
		files["tests/handler_test.go"] = templates.HandlerTestTemplate(data.Name, data.ModuleType)
		files["tests/service_test.go"] = templates.ServiceTestTemplate(data.Name, data.ModuleType)
	}

	// Write files
	for filePath, tmpl := range files {
		if err := utils.WriteTemplateFile(
			filepath.Join(basePath, filePath),
			tmpl,
			data,
		); err != nil {
			return fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
	}

	fmt.Printf("Successfully generated %s '%s'\n", data.ModuleType, data.Name)
	return nil
}
