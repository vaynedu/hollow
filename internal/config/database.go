package config

type DbConfig struct {
	DSN     string `mapstructure:"dsn"`
	Dialect string `mapstructure:"dialect"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}
