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
}

func New(cfg config.Config) *Manager {
	return &Manager{
		Cfg: cfg,
		Msg: locale.SelectMessages(cfg.Locale),
	}
}
