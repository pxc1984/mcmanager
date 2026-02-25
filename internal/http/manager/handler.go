package manager

import (
	"log"

	"github.com/gin-gonic/gin"
)

func (m *Manager) UpdateHandler(c *gin.Context) {
	if m.Cfg.SecretToken != "" {
		token := c.GetHeader("X-Secret-Token")
		if token == "" || token != m.Cfg.SecretToken {
			log.Printf("unauthorized update request from %s", c.ClientIP())
			c.String(401, "unauthorized")
			return
		}
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()

	log.Printf("received update request from %s", c.ClientIP())

	if err := m.syncRepoFn(); err != nil {
		log.Printf("repo sync failed: %v", err)
		c.String(500, "repo sync failed: %v", err)
		return
	}

	if err := m.pluginDownloadFn(); err != nil {
		log.Printf("plugin download failed: %v", err)
		c.String(500, "plugin download failed: %v", err)
		return
	}

	if err := m.dataSyncFn(); err != nil {
		log.Printf("data sync failed: %v", err)
		c.String(500, "data sync failed: %v", err)
		return
	}

	go func() {
		if err := m.restartFn(); err != nil {
			log.Printf("restart failed: %v", err)
		}
	}()

	c.String(200, "update applied; restart scheduled")
}
