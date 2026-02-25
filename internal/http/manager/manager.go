package manager

import (
	"mcmanager/internal/config"
	"mcmanager/internal/locale"
	"sync"
)

type Manager struct {
	Cfg config.Config
	Mu  sync.Mutex
	Msg locale.Messages

	syncRepoFn       func() error
	pluginDownloadFn func() error
	dataSyncFn       func() error
	restartFn        func() error
}

func New(cfg config.Config) *Manager {
	m := &Manager{
		Cfg: cfg,
		Msg: locale.SelectMessages(cfg.Locale),
	}
	m.syncRepoFn = m.syncRepo
	m.pluginDownloadFn = m.runPluginDownload
	m.dataSyncFn = m.syncData
	m.restartFn = m.restartWithCountdown
	return m
}
