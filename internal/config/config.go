package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	c        *viper.Viper
	confPath string
)

// Config is a struct representing the application's configuration.
type Config struct {
	AppName    string
	RunMode    string
	Protocol   string
	Domain     string
	Host       string
	Port       int
	JWTSecret  string
	JWTExpTime int
	Logger     Logger
	Data       Data
	Github     Github
	Facebook   Facebook
	AWS        AWS
	Mailgun    Mailgun
}

func init() {
	flag.StringVar(&confPath, "conf", "", "e.g: bin ./config.yaml")
}

// Init initializes and loads the application configuration.
func Init() (*Config, error) {
	flag.Parse()
	return load(confPath)
}

func load(in string) (*Config, error) {
	c = viper.New()

	// Add the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	c.SetConfigFile(in)
	// By default, read from config.{yaml,toml, yml,json}, etc. files
	c.AddConfigPath(in)
	c.AddConfigPath("/etc/stocms")
	c.AddConfigPath("$HOME/.stocms")
	c.AddConfigPath(".")
	c.AddConfigPath(filepath.Dir(ex))

	err = c.ReadInConfig()

	return &Config{
		AppName:    c.GetString("app_name"),
		RunMode:    c.GetString("run_mode"),
		Protocol:   c.GetString("server.protocol"),
		Domain:     c.GetString("server.domain"),
		Host:       c.GetString("server.host"),
		Port:       c.GetInt("server.port"),
		JWTSecret:  c.GetString("jwt.secret"),
		JWTExpTime: c.GetInt("jwt.exp_time"),
		Logger:     *getLog(),
		Data: Data{
			Database: *getDatabase(),
			Redis:    *getRedis(),
		},
		Github:   *getGithub(),
		Facebook: *getFacebook(),
		AWS:      *getAWS(),
		Mailgun:  *getMailgun(),
	}, err
}
