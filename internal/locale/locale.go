package locale

type Messages struct {
	RestartSoon string
	Countdown   string
}

func SelectMessages(locale string) Messages {
	switch locale {
	case "ru":
		return Messages{
			RestartSoon: "Обновление получено, перезапуск через 60 секунд",
			Countdown:   "Перезапуск через %d",
		}
	default:
		return Messages{
			RestartSoon: "Update received, restarting in 60 seconds",
			Countdown:   "Restarting in %d",
		}
	}
}
