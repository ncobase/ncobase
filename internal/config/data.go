package config

import "time"

// Data data config struct
type Data struct {
	Database Database
	Redis    Redis
}

// Database database config struct
type Database struct {
	Driver          string
	Source          string
	Migrate         bool
	TablePrefix     string
	MaxIdleConn     int
	MaxOpenConn     int
	ConnMaxLifeTime time.Duration
}

// Redis redis config struct
type Redis struct {
	Addr         string
	Username     string
	Password     string
	Db           int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
}

func getDataConfig() Data {
	return Data{
		Database: Database{
			Driver:          c.GetString("data.database.driver"),
			Source:          c.GetString("data.database.source"),
			Migrate:         c.GetBool("data.database.migrate"),
			TablePrefix:     c.GetString("data.database.table_prefix"),
			MaxIdleConn:     c.GetInt("data.database.max_idle_conn"),
			MaxOpenConn:     c.GetInt("data.database.max_open_conn"),
			ConnMaxLifeTime: c.GetDuration("data.database.max_life_time"),
		},
		Redis: Redis{
			Addr:         c.GetString("data.redis.addr"),
			Username:     c.GetString("data.redis.username"),
			Password:     c.GetString("data.redis.password"),
			Db:           c.GetInt("data.redis.db"),
			ReadTimeout:  c.GetDuration("data.redis.read_timeout"),
			WriteTimeout: c.GetDuration("data.redis.write_timeout"),
			DialTimeout:  c.GetDuration("data.redis.dial_timeout"),
		},
	}
}
