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
	UseMongo bool
	UseEnt   bool
	UseGorm  bool
	WithTest bool
	Group    string
}

var extDescriptions = map[string]string{
	"core":   "Core Domain",
	"domain": "Business Domain",
	"plugin": "Plugin Domain",
}

// Generate generates code based on options
func Generate(opts *Options) error {
	if !utils.ValidateName(opts.Name) {
		return fmt.Errorf("invalid name: %s", opts.Name)
	}

	var basePath string
	var extType string
	var mainTemplate func(string) string

	switch opts.Type {
	case "core":
		basePath = filepath.Join("core", opts.Name)
		extType = "core"
		mainTemplate = templates.CoreTemplate
	case "business":
		basePath = filepath.Join("domain", opts.Name)
		extType = "domain"
		mainTemplate = templates.BusinessTemplate
	case "plugin":
		basePath = filepath.Join("plugin", opts.Name)
		extType = "plugin"
		mainTemplate = templates.PluginTemplate
	default:
		return fmt.Errorf("unknown type: %s", opts.Type)
	}

	// Check if component already exists
	if exists, err := utils.PathExists(basePath); err != nil {
		return fmt.Errorf("error checking existence: %v", err)
	} else if exists {
		return fmt.Errorf("'%s' already exists in %s", opts.Name, extDescriptions[extType])
	}

	// Prepare template data
	data := &templates.Data{
		Name:        opts.Name,
		Type:        opts.Type,
		UseMongo:    opts.UseMongo,
		UseEnt:      opts.UseEnt,
		UseGorm:     opts.UseGorm,
		WithTest:    opts.WithTest,
		Group:       opts.Group,
		ExtType:     extType,
		PackagePath: fmt.Sprintf("ncobase/%s/%s", extType, opts.Name),
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
			return templates.DataTemplateWithEnt(data.Name, data.ExtType)
		}
		if data.UseGorm {
			return templates.DataTemplateWithGorm(data.Name, data.ExtType)
		}
		if data.UseMongo {
			return templates.DataTemplateWithMongo(data.Name, data.ExtType)
		}
		return templates.DataTemplate(data.Name, data.ExtType)
	}

	files := map[string]string{
		fmt.Sprintf("%s.go", data.Name): mainTemplate(data.Name),
		"data/data.go":                  selectDataTemplate(*data),
		"data/repository/provider.go":   templates.RepositoryTemplate(data.Name, data.ExtType),
		"data/schema/schema.go":         templates.SchemaTemplate(),
		"handler/provider.go":           templates.HandlerTemplate(data.Name, data.ExtType),
		"service/provider.go":           templates.ServiceTemplate(data.Name, data.ExtType),
		"structs/structs.go":            templates.StructsTemplate(),
	}

	// Add ent files if required
	if data.UseEnt {
		files["generate.go"] = templates.GeneraterTemplate(data.Name, data.ExtType)
	}

	// Add test files if required
	if data.WithTest {
		files["tests/ext_test.go"] = templates.ExtTestTemplate(data.Name, data.ExtType)
		files["tests/handler_test.go"] = templates.HandlerTestTemplate(data.Name, data.ExtType)
		files["tests/service_test.go"] = templates.ServiceTestTemplate(data.Name, data.ExtType)
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

	fmt.Printf("Successfully generated '%s' in %s\n", data.Name, extDescriptions[data.ExtType])
	return nil
}
