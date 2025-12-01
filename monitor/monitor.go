package monitor

import (
	"ANFeGuard/database"
	"ANFeGuard/email"
	"ANFeGuard/sysinfo"
	"ANFeGuard/winservice"
	"fmt"
	"log"
	"time"
)

type MonitorConfig struct {
	Period      time.Duration
	Services    []string
	EmailConfig email.SMTPConfig
	Recipients  []string
	MaxRetries  int
	CPULimit    float64
	MemLimit    float64
	DiskPath    string
}

// Start inicia o monitoramento unificado (serviços + recursos)
func Start(cfg MonitorConfig) {
	fmt.Println("Monitor ANFeGuard iniciado — intervalo:", cfg.Period)

	ticker := time.NewTicker(cfg.Period)
	defer ticker.Stop()

	for range ticker.C {
		monitorSystem(cfg)
		monitorServices(cfg)
	}
}

// ========== MONITORAMENTO DE SISTEMA ==========
func monitorSystem(cfg MonitorConfig) {
	info, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		log.Println("[ERRO] Coleta de sistema:", err)
		return
	}

	currentTime := time.Now().Format("02/01/2006 15:04:05")

	if info.CPUPercent > cfg.CPULimit || info.MemoryPercent > cfg.MemLimit {

		data := email.EmailAlertData{
			Service: "", // não é alerta de serviço
			CPU:     info.CPUPercent,
			Memory:  info.MemoryPercent,
			Disk:    info.DiskUsedPercent,
			Time:    currentTime,
		}

		err := email.SendEmail(
			cfg.EmailConfig,
			cfg.Recipients,
			"🚨 Alerta ANFeGuard — Uso elevado de recursos",
			"email/templates/alerta.html",
			data,
		)

		if err != nil {
			log.Println("[ERRO] Envio de e-mail de alerta de sistema:", err)
		}
	}
}

// ========== MONITORAMENTO DE SERVIÇOS ==========
func monitorServices(cfg MonitorConfig) {
	sysInfo, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		log.Println("[ERRO] Falha ao coletar informações do sistema:", err)
		return
	}
	for _, svc := range cfg.Services {
		status, err := winservice.GetStatus(svc)
		if err != nil {
			msg := fmt.Sprintf("Falha ao obter status: %v", err)
			log.Printf("[ERRO] Serviço '%s': %v\n", svc, err)
			database.LogServiceError(svc, "Unknown", msg, sysInfo.MemoryPercent)
			continue

		}

		if status == winservice.StatusStopped {
			log.Printf("[ALERTA] Serviço '%s' está parado. Tentando iniciar...\n", svc)
			msg := "Serviço parado"

			database.LogServiceError(svc, string(status), msg, sysInfo.MemoryPercent)
			if err := winservice.Start(svc); err != nil {
				msg := "Serviço parado"
				log.Printf("[ERRO] Falha ao iniciar '%s': %v\n", svc, err)
				sendServiceEmail(cfg, svc, "Falha ao iniciar serviço")
				database.LogServiceError(svc, string(status), msg, sysInfo.MemoryPercent)
				continue
			}

			log.Printf("[OK] Serviço '%s' inciado com sucesso.\n", svc)
			sendServiceEmail(cfg, svc, "Serviço iniciado automaticamente")
		}
	}
}

func sendServiceEmail(cfg MonitorConfig, serviceName, subject string) {

	sysInfo, _ := sysinfo.GetSystemInfo(cfg.DiskPath)

	data := email.EmailAlertData{
		Service: serviceName,
		CPU:     sysInfo.CPUPercent,
		Memory:  sysInfo.MemoryPercent,
		Disk:    sysInfo.DiskUsedPercent,
		Time:    time.Now().Format("02/01/2006 15:04:05"),
	}

	template := "email/templates/service_stopped.html"

	if err := email.SendEmail(cfg.EmailConfig, cfg.Recipients, subject, template, data); err != nil {
		log.Printf("[ERRO] Falha ao enviar e-mail do serviço '%s': %v\n", serviceName, err)
	}
}
