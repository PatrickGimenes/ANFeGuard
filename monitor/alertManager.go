package monitor

import (
	"ANFeGuard/logs"
	"sync"
	"time"
)

var lastAlertTime time.Time
var alertCooldown = 30 * time.Minute

var highSentMutex sync.Mutex

func HandleResourceAlert(data ServerData, cfg MonitorConfig) {
	highSentMutex.Lock()
	defer highSentMutex.Unlock()

	isHighUsage := data.CPU > cfg.CPULimit || data.Memory > cfg.MemLimit
	now := time.Now()

	if isHighUsage {
		if now.Sub(lastAlertTime) > alertCooldown {
			logs.Warn("Limites excedidos, enviando alerta")

			sendServiceEmail(cfg, "", "ResourceAlert", "Alerta ANFeGuard — Uso elevado de recursos")

			lastAlertTime = now
		} else {
			logs.Warn("Já foi enviado um e-mail sobre alto consumo, não enviando novo e-mail")
		}
		return
	}

	logs.Info("Uso dos recursos normalizado")
}
