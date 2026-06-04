package config

const (
	ConfigBasePath = ".build/config/local.yaml"
)

type Env = string

const (
	EnvLocal      Env = "local"
	EnvDebug      Env = "debug"
	EnvProduction Env = "production"
)
