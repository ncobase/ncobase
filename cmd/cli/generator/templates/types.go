package templates

// Data represents template data structure
type Data struct {
	Name        string // Extension name
	Type        string // Extension type (core/business/plugin)
	UseMongo    bool   // Whether to use MongoDB
	UseEnt      bool   // Whether to use Ent ORM
	UseGorm     bool   // Whether to use GORM
	WithTest    bool   // Whether to generate test files
	Group       string // Optional group name
	ExtType     string // Extension type in belongs domain path (core/domain/plugin)
	PackagePath string // Full package path
}
