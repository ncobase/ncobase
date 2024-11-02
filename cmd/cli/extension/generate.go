package extension

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ncobase/cmd/cli/extension/templates"

	"github.com/spf13/cobra"
)

const (
	ErrEmptyName     = "Error: Name cannot be empty. Please provide a name for your core, business or plugin."
	ErrInvalidName   = "Error: Invalid name. Use only alphanumeric characters, underscores, and hyphens."
	ErrInvalidType   = "Error: Invalid generation type. Use 'core', 'business' or 'plugin'."
	ErrAlreadyExists = "Error: '%s' already exists in %s. Please choose a different name or remove the existing one."
)

// Cmd is the main generate command
var Cmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"gen", "generate"},
	Short:   "Generate new components",
	Long:    `Generate new extensions, plugins, or other components for Ncobase.`,
}

var genCoreCmd = &cobra.Command{
	Use:     "core [name]",
	Aliases: []string{"c"},
	Short:   "Generate a new core extension",
	Long:    `Generate a new core extension with the specified name.`,
	Args:    cobra.ExactArgs(1),
	Run:     runGenerate,
}

var genBusinessCmd = &cobra.Command{
	Use:     "business [name]",
	Aliases: []string{"b"},
	Short:   "Generate a new business extension",
	Long:    `Generate a new business extension with the specified name.`,
	Args:    cobra.ExactArgs(1),
	Run:     runGenerate,
}

var genPluginCmd = &cobra.Command{
	Use:     "plugin [name]",
	Aliases: []string{"p"},
	Short:   "Generate a new plugin",
	Long:    `Generate a new plugin with the specified name.`,
	Args:    cobra.ExactArgs(1),
	Run:     runGenerate,
}

func init() {
	Cmd.AddCommand(genCoreCmd, genBusinessCmd, genPluginCmd)
}

func runGenerate(cmd *cobra.Command, args []string) {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		fmt.Println(ErrEmptyName)
		os.Exit(1)
	}

	name := strings.TrimSpace(args[0])

	if !isValidName(name) {
		fmt.Println(ErrInvalidName)
		os.Exit(1)
	}

	var basePath string
	var generateFunc func(string)

	switch cmd.CalledAs() {
	case "core", "c":
		basePath = filepath.Join("core", name)
		generateFunc = generateCore
	case "business", "b":
		basePath = filepath.Join("domain", name)
		generateFunc = generateBusiness
	case "plugin", "p":
		basePath = filepath.Join("plugin", name)
		generateFunc = generatePlugin
	default:
		fmt.Println(ErrInvalidType)
		os.Exit(1)
	}

	// Check if the extension/plugin already exists
	if exists, err := pathExists(basePath); err != nil {
		fmt.Printf("Error checking existence: %v\n", err)
		os.Exit(1)
	} else if exists {
		fmt.Printf(ErrAlreadyExists, name, filepath.Dir(basePath))
		os.Exit(1)
	}

	generateFunc(name)
}

// isValidName checks if the provided name contains only allowed characters
func isValidName(name string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	return matched
}

// pathExists checks if a file or directory exists
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func generateCore(name string) {
	basePath := filepath.Join("core", name)
	createStructure(basePath, name, templates.CoreTemplate, "core")
}

func generateBusiness(name string) {
	basePath := filepath.Join("domain", name)
	createStructure(basePath, name, templates.BusinessTemplate, "domain")
}

func generatePlugin(name string) {
	basePath := filepath.Join("plugin", name)
	createStructure(basePath, name, templates.PluginTemplate, "plugin")
}

func createStructure(basePath, name string, mainTemplate func(string) string, moduleType string) {
	if err := createDirectory(basePath); err != nil {
		fmt.Printf("Error creating base directory: %v\n", err)
		os.Exit(1)
	}

	directories := []string{
		"data",
		"data/repository",
		"data/schema",
		"handler",
		"service",
		"structs",
	}

	for _, dir := range directories {
		if err := createDirectory(filepath.Join(basePath, dir)); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	files := map[string]string{
		fmt.Sprintf("%s.go", name):      mainTemplate(name),
		"data/data.go":                  templates.DataTemplate(name, moduleType),
		"data/repository/repository.go": templates.RepositoryTemplate(name, moduleType),
		"data/schema/schema.go":         templates.SchemaTemplate(),
		"handler/handler.go":            templates.HandlerTemplate(name, moduleType),
		"service/service.go":            templates.ServiceTemplate(name, moduleType),
		"structs/structs.go":            templates.StructsTemplate(),
	}

	for filePath, content := range files {
		if err := createFile(filepath.Join(basePath, filePath), content); err != nil {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
			os.Exit(1)
		}
	}

	fmt.Printf("%s '%s' generated successfully.\n", moduleType, name)
}

func createDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func createFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
