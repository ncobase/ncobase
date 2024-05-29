package config

// Logger logger config struct
type Logger struct {
	Level      int
	Path       string
	Format     string
	Output     string
	OutputFile string
}

func getLog() *Logger {
	return &Logger{
		Level:      c.GetInt("logger.level"),
		Format:     c.GetString("logger.format"),
		Path:       c.GetString("logger.path"),
		Output:     c.GetString("logger.output"),
		OutputFile: c.GetString("logger.output_file"),
	}
}
