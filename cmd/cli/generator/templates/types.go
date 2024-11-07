package templates

// Data represents template data structure
type Data struct {
	Name        string // Component name
	Type        string // Component type (core/module/plugin)
	Group       string // Optional group name
	UseEnt      bool   // Whether to use Ent ORM
	UseGorm     bool   // Whether to use GORM
	WithTest    bool   // Whether to generate test files
	ExtType     string // extension type in path (core/domain/plugin)
	PackagePath string // Full package path
}
