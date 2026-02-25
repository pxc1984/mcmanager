package manager

import (
	"fmt"
	"time"

	"github.com/gorcon/rcon"
)

func (m *Manager) restartWithCountdown() error {
	address := fmt.Sprintf("%s:%d", m.Cfg.RconHost, m.Cfg.RconPort)
	client, err := rcon.Dial(address, m.Cfg.RconPassword)
	if err != nil {
		return fmt.Errorf("connect to rcon: %w", err)
	}
	defer client.Close()

	announce := func(msg string) error {
		_, err := client.Execute(fmt.Sprintf("say %s", msg))
		return err
	}

	if err := announce(m.Msg.RestartSoon); err != nil {
		return fmt.Errorf("announce restart: %w", err)
	}

	time.Sleep(time.Second * time.Duration(m.Cfg.CountdownWait))

	for i := 10; i >= 1; i-- {
		if err := announce(fmt.Sprintf(m.Msg.Countdown, i)); err != nil {
			return fmt.Errorf("countdown announce: %w", err)
		}
		time.Sleep(1 * time.Second)
	}

	if _, err := client.Execute(m.Cfg.RestartCmd); err != nil {
		return fmt.Errorf("send restart command: %w", err)
	}

	return nil
}
