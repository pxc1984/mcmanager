package manager

import (
	"log"

	"github.com/gin-gonic/gin"
)

func (m *Manager) UpdateHandler(c *gin.Context) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	log.Printf("received update request from %s", c.ClientIP())

	if err := m.syncRepo(); err != nil {
		log.Printf("repo sync failed: %v", err)
		c.String(500, "repo sync failed: %v", err)
		return
	}

	if err := m.runPluginDownload(); err != nil {
		log.Printf("plugin download failed: %v", err)
		c.String(500, "plugin download failed: %v", err)
		return
	}

	if err := m.syncData(); err != nil {
		log.Printf("data sync failed: %v", err)
		c.String(500, "data sync failed: %v", err)
		return
	}

	go func() {
		if err := m.restartWithCountdown(); err != nil {
			log.Printf("restart failed: %v", err)
		}
	}()

	c.String(200, "update applied; restart scheduled")
}
