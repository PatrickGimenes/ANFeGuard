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
	EmailConfig email.SMTPConfig
	Recipients  []string
	MaxRetries  int
	CPULimit    float64
	MemLimit    float64
	DiskPath    string
}

// Start inicia o monitoramento unificado (serviços + recursos)
func Start(cfg MonitorConfig) {
	log.Println("[INFO] Monitor ANFeGuard iniciado — intervalo:", cfg.Period)

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

	fmt.Printf("Horário: %v | CPU: %.1f%% | Memória: %.1f%% | Disco: %.1f%%\n", currentTime,
		info.CPUPercent, info.MemoryPercent, info.DiskUsedPercent)

	if info.CPUPercent > cfg.CPULimit || info.MemoryPercent > cfg.MemLimit {

		// data := email.EmailAlertData{
		// 	Service: "", // não é alerta de serviço
		// 	CPU:     fmt.Sprintf("%.2f%%", info.CPUPercent),
		// 	Memory:  fmt.Sprintf("%.2f%%", info.MemoryPercent),
		// 	Disk:    fmt.Sprintf("%.2f%%", info.DiskUsedPercent),
		// 	DiskPath: cfg.DiskPath,
		// 	Time:    currentTime,
		// }

		// err := email.SendEmail(
		// 	cfg.EmailConfig,
		// 	cfg.Recipients,
		// 	"🚨 Alerta ANFeGuard — Uso elevado de recursos",
		// 	"email/templates/alerta.html",
		// 	data,
		// )
		sendServiceEmail(cfg, "", "🚨 Alerta ANFeGuard — Uso elevado de recursos")
	}
}

// ========== MONITORAMENTO DE SERVIÇOS ==========
func monitorServices(cfg MonitorConfig) {
	services := database.GetServices()
	tries := 0
	sysInfo, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		log.Println("[ERRO] Falha ao coletar informações do sistema:", err)
		return
	}
	for _, svc := range services {
		status, err := winservice.GetStatus(svc)
		if err != nil {
			msg := fmt.Sprintf("Falha ao obter status: %v", err)
			log.Printf("[ERRO] Serviço '%s': %v\n", svc, err)
			database.LogServiceError(svc, "Unknown", msg, sysInfo.MemoryPercent)
			continue

		}

		if status == winservice.StatusStopped {

			if (tries <= cfg.MaxRetries){
			log.Printf("[ALERTA] Serviço '%s' está parado. Tentando iniciar...\n", svc)
			msg := "Serviço parado"
			tries++

			sendServiceEmail(cfg, svc, string(status), "Serviço parado. Tentando iniciar...")
			database.LogServiceError(svc, string(status), msg, sysInfo.MemoryPercent)

			if err := winservice.Start(svc); err != nil {
				msg := "Falha ao iniciar o serviço"
				log.Printf("[ERRO] Falha ao iniciar '%s': %v\n", svc, err)
				sendServiceEmail(cfg, svc, string(status), "Falha ao iniciar serviço!")
				database.LogServiceError(svc, string(status), msg, sysInfo.MemoryPercent)
				continue
			}

			log.Printf("[SUCESSO] Serviço '%s' inciado com sucesso.\n", svc)
			sendServiceEmail(cfg, svc, "Serviço iniciado com sucesso!")
			tries = 0
		}else{
			log.Printf("[ERRO] Serviço '%s' alcançou o maxímo de tentativas.\n", svc)
			sendServiceEmail(cfg, svc, "Serviço alcançou o maxímo de tentativas.")
		}
	}
	}
}

func sendServiceEmail(cfg MonitorConfig, serviceName, status string, subject string) {

	sysInfo, _ := sysinfo.GetSystemInfo(cfg.DiskPath)
	data := email.EmailAlertData{
		Service: serviceName,
		CPU:     fmt.Sprintf("%.2f%%", sysInfo.CPUPercent),
		Memory:  fmt.Sprintf("%.2f%%", sysInfo.MemoryPercent),
		Disk:    fmt.Sprintf("%.2f%%", sysInfo.DiskUsedPercent),
		DiskPath: cfg.DiskPath,
		Time:    time.Now().Format("02/01/2006 15:04:05"),
	}
	
	templatePath := ""
	if status == "Stopped"{
		templatePath = "email/templates/service_stopped.html"
	}else{
		templatePath = "email/templates/service_started.html"
	}

	template := templatePath

	if err := email.SendEmail(cfg.EmailConfig, cfg.Recipients, subject, template, data); err != nil {
		log.Printf("[ERRO] Falha ao enviar e-mail do serviço '%s': %v\n", serviceName, err)
	}
}
