package feature

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ncobase/cmd/cli/feature/templates"

	"github.com/spf13/cobra"
)

// Cmd is the main generate command
var Cmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"gen", "generate"},
	Short:   "Generate new components",
	Long:    `Generate new features, plugins, or other components for Ncobase.`,
}

var genFeatureCmd = &cobra.Command{
	Use:     "feature [name]",
	Aliases: []string{"f"},
	Short:   "Generate a new feature",
	Long:    `Generate a new feature with the specified name.`,
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
	Cmd.AddCommand(genFeatureCmd, genPluginCmd)
}

func runGenerate(cmd *cobra.Command, args []string) {
	if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
		fmt.Println("Error: Name cannot be empty. Please provide a name for your feature or plugin.")
		os.Exit(1)
	}

	name := strings.TrimSpace(args[0])

	if !isValidName(name) {
		fmt.Println("Error: Invalid name. Use only alphanumeric characters, underscores, and hyphens.")
		os.Exit(1)
	}

	var basePath string
	var generateFunc func(string)

	switch cmd.CalledAs() {
	case "feature", "f":
		basePath = filepath.Join("feature", name)
		generateFunc = generateFeature
	case "plugin", "p":
		basePath = filepath.Join("plugin", name)
		generateFunc = generatePlugin
	default:
		fmt.Println("Invalid generation type. Use 'feature' or 'plugin'.")
		os.Exit(1)
	}

	// Check if the feature/plugin already exists
	if exists, err := pathExists(basePath); err != nil {
		fmt.Printf("Error checking existence: %v\n", err)
		os.Exit(1)
	} else if exists {
		fmt.Printf("Error: '%s' already exists in %s. Please choose a different name or remove the existing one.\n", name, filepath.Dir(basePath))
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

func generateFeature(name string) {
	basePath := filepath.Join("feature", name)
	createStructure(basePath, name, templates.FeatureMainTemplate, "feature")
}

func generatePlugin(name string) {
	basePath := filepath.Join("plugin", name)
	createStructure(basePath, name, templates.PluginMainTemplate, "plugin")
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
