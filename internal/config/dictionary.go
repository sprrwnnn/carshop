package config

import "time"

type (
	Config struct {
		Env         string   `yaml:"env" env-default:"production"`
		Server      Server   `yaml:"server"`
		ValkeyCache Cache    `yaml:"cache"`
		Database    Database `yaml:"-"`
		RabbitMQ    RabbitMQ `yaml:"rabbitmq"`
		Build       BuildInfo
	}

	BuildInfo struct {
		Version string `env:"APP_VERSION"`
		Commit  string `env:"APP_COMMIT"`
		Date    string `env:"APP_BUILD_DATE"`
		Name    string `env:"APP_NAME"`
	}

	Server struct {
		Host string `yaml:"host" env-default:"localhost"`
		Port int    `yaml:"port" env-default:"8000"`
	}

	Cache struct {
		Addrs       string        `yaml:"-" env:"CACHE_ADDRS" env-default:"localhost:6379"`
		Password    string        `yaml:"-" env:"CACHE_PASSWORD" env-default:""`
		ReadOnly    bool          `yaml:"readOnly" env-default:"false"`
		DialTimeout time.Duration `yaml:"dialTimeout" env-default:"5s"`
		PoolSize    int           `yaml:"poolSize" env-default:"10"`
		DefaultTTL  time.Duration `yaml:"defaultTtl" env-default:"1h"`
	}

	Database struct {
		Hosts    string `yaml:"-" env:"DB_HOSTS"`
		Port     uint16 `yaml:"-" env:"DB_PORT"`
		Username string `yaml:"-" env:"DB_USERNAME"`
		Password string `yaml:"-" env:"DB_PASSWORD"`
		DBName   string `yaml:"-" env:"DB_DBNAME"`
	}

	RabbitMQ struct {
		URL string `yaml:"-" env:"RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
	}
)

type (
	LocalDeploy struct {
		Deploy LocalDeployConfig `yaml:"deploy"`
	}

	LocalDeployConfig struct {
		Env []LocalDeployConfigElement `yaml:"env"`
	}

	LocalDeployConfigElement struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	}
)
