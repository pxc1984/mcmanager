package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	RepoURL         string `env:"REPO_URL"`
	RepoBranch      string `env:"REPO_BRANCH" envDefault:"main"`
	RepoPath        string `env:"REPO_PATH" envDefault:"/tmp/plugin-repo"`
	DataDir         string `env:"DATA_DIR" envDefault:"./data"`
	Port            string `env:"PORT" envDefault:"8080"`
	RconHost        string `env:"RCON_HOST"`
	RconPort        int    `env:"RCON_PORT"`
	RconPassword    string `env:"RCON_PASSWORD"`
	RestartCmd      string `env:"RCON_RESTART_COMMAND" envDefault:"restart"`
	CountdownWait   int    `env:"CountdownWait" envDefault:"50"`
	PluginsUID      int    `env:"PLUGINS_UID" envDefault:"0"`
	PluginsDownload bool   `env:"PLUGINS_DOWNLOAD" envDefault:"true"`
	SecretToken     string `env:"SECRET_TOKEN" envDefault:""`
	Locale          string `env:"LOCALE" envDefault:"en"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
