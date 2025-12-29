package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RepoURL       string
	RepoBranch    string
	RepoPath      string
	DataDir       string
	HTTPAddr      string
	RconHost      string
	RconPort      int
	RconPassword  string
	RestartCmd    string
	CountdownWait time.Duration
	PluginsUID    int
	HasPluginsUID bool
	Locale        string
}

func Load() (Config, error) {
	repoURL := os.Getenv("REPO_URL")
	if repoURL == "" {
		return Config{}, errors.New("REPO_URL is required")
	}

	repoBranch := os.Getenv("REPO_BRANCH")
	if repoBranch == "" {
		repoBranch = "main"
	}

	repoPath := os.Getenv("REPO_PATH")
	if repoPath == "" {
		repoPath = "/tmp/plugin-repo"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	rconHost := os.Getenv("RCON_HOST")
	if rconHost == "" {
		return Config{}, errors.New("RCON_HOST is required")
	}

	rconPortStr := os.Getenv("RCON_PORT")
	if rconPortStr == "" {
		return Config{}, errors.New("RCON_PORT is required")
	}
	rconPort, err := strconv.Atoi(rconPortStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid RCON_PORT: %w", err)
	}

	rconPassword := os.Getenv("RCON_PASSWORD")
	if rconPassword == "" {
		return Config{}, errors.New("RCON_PASSWORD is required")
	}

	restartCmd := os.Getenv("RCON_RESTART_COMMAND")
	if restartCmd == "" {
		restartCmd = "restart"
	}

	locale := os.Getenv("LOCALE")
	if locale == "" {
		locale = "en"
	}

	pluginsUIDStr := os.Getenv("PLUGINS_UID")
	var pluginsUID int
	hasPluginsUID := false
	if pluginsUIDStr != "" {
		uid, err := strconv.Atoi(pluginsUIDStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid PLUGINS_UID: %w", err)
		}
		pluginsUID = uid
		hasPluginsUID = true
	}

	return Config{
		RepoURL:       repoURL,
		RepoBranch:    repoBranch,
		RepoPath:      repoPath,
		DataDir:       dataDir,
		HTTPAddr:      ":" + port,
		RconHost:      rconHost,
		RconPort:      rconPort,
		RconPassword:  rconPassword,
		RestartCmd:    restartCmd,
		CountdownWait: 50 * time.Second,
		PluginsUID:    pluginsUID,
		HasPluginsUID: hasPluginsUID,
		Locale:        locale,
	}, nil
}
