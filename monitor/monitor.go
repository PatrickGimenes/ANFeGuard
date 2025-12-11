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

// Controla tentativas por serviço
var retryCount = map[string]int{}

// =====================================================
// INICIO DO MONITORAMENTO
// =====================================================
func Start(cfg MonitorConfig) {
	log.Printf("[INFO] ANFeGuard Monitor iniciado | Intervalo: %s\n", cfg.Period)

	ticker := time.NewTicker(cfg.Period)
	defer ticker.Stop()

	for range ticker.C {
		monitorSystem(cfg)
		monitorServices(cfg)
	}
}

// =====================================================
// MONITORAMENTO DE SISTEMA (CPU / RAM / DISCO)
// =====================================================
func monitorSystem(cfg MonitorConfig) {
	info, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		log.Printf("[ERROR] Falha ao coletar informações de sistema: %v\n", err)
		return
	}

	now := time.Now().Format("02/01/2006 15:04:05")

	log.Printf("[INFO] %s | CPU: %.1f%% | RAM: %.1f%% | Disco(%s): %.1f%%",
		now, info.CPUPercent, info.MemoryPercent, cfg.DiskPath, info.DiskUsedPercent)

	// Verifica limites
	if info.CPUPercent > cfg.CPULimit || info.MemoryPercent > cfg.MemLimit {
		log.Printf("[ALERT] Limites de recursos excedidos (CPU/RAM)\n")
		sendServiceEmail(cfg, "", "ResourceAlert", "Alerta ANFeGuard — Uso elevado de recursos")
	}
}

// =====================================================
// MONITORAMENTO DE SERVIÇOS
// =====================================================
func monitorServices(cfg MonitorConfig) {
	services := database.GetServices()
	sysInfo, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		log.Printf("[ERROR] Falha ao coletar system info para serviços: %v\n", err)
		return
	}

	for _, svc := range services {
		status, err := winservice.GetStatus(svc)
		if err != nil {
			logServiceError(svc, "Unknown", fmt.Sprintf("Erro ao obter status: %v", err), &sysInfo)
			continue
		}

		if status != winservice.StatusStopped {
			resetRetries(svc)
			continue
		}

		// Serviço parado
		retryServiceStart(cfg, svc, status, &sysInfo)
	}
}

// =====================================================
// LÓGICA DE RETENTATIVA DE INÍCIO DE SERVIÇO
// =====================================================
func retryServiceStart(cfg MonitorConfig, svc string, status winservice.Status, sysInfo *sysinfo.SysInfo) {

	retryCount[svc]++

	// Excedeu tentativas
	if retryCount[svc] > cfg.MaxRetries {
		log.Printf("[ERROR] Serviço '%s' atingiu o máximo de tentativas (%d)\n", svc, cfg.MaxRetries)
		sendServiceEmail(cfg, svc, "MaxRetries", "ANFeGuard — Máximo de tentativas atingido")
		return
	}

	// Tentando iniciar
	log.Printf("[ALERT] Serviço '%s' está parado. Tentativa %d/%d\n",
		svc, retryCount[svc], cfg.MaxRetries)

	logServiceError(svc, string(status), "Serviço parado", sysInfo)

	sendServiceEmail(cfg, svc, "Stopped", "Serviço parado — Tentando iniciar...")

	// Tentar iniciar
	if err := winservice.Start(svc); err != nil {
		msg := fmt.Sprintf("Falha ao iniciar: %v", err)
		log.Printf("[ERROR] %s\n", msg)

		sendServiceEmail(cfg, svc, "StartFailed", "Falha ao iniciar serviço!")
		logServiceError(svc, string(status), msg, sysInfo)

		return
	}

	// Sucesso — reseta tentativas
	resetRetries(svc)

	log.Printf("[SUCCESS] Serviço '%s' iniciado com sucesso.\n", svc)
	sendServiceEmail(cfg, svc, "Started", "Serviço iniciado com sucesso!")
}

// =====================================================
// RESET DE RETENTATIVAS
// =====================================================
func resetRetries(service string) {
	if retryCount[service] > 0 {
		log.Printf("[INFO] Resetando tentativas do serviço '%s'\n", service)
	}
	retryCount[service] = 0
}

// =====================================================
// LOG DE ERRO CENTRALIZADO
// =====================================================
func logServiceError(svc, status, msg string, info *sysinfo.SysInfo) {
	log.Printf("[ERROR] Serviço '%s' | Status: %s | %s\n", svc, status, msg)
	database.LogServiceError(svc, status, msg, info.MemoryPercent)
}

// =====================================================
// ENVIO DE EMAIL CENTRALIZADO
// =====================================================
func sendServiceEmail(cfg MonitorConfig, serviceName, status string, subject string) {

	sysInfo, _ := sysinfo.GetSystemInfo(cfg.DiskPath)

	data := email.EmailAlertData{
		Service:  serviceName,
		CPU:      fmt.Sprintf("%.2f%%", sysInfo.CPUPercent),
		Memory:   fmt.Sprintf("%.2f%%", sysInfo.MemoryPercent),
		Disk:     fmt.Sprintf("%.2f%%", sysInfo.DiskUsedPercent),
		DiskPath: cfg.DiskPath,
		Time:     time.Now().Format("02/01/2006 15:04:05"),
	}

	template := selectTemplate(status)

	if err := email.SendEmail(cfg.EmailConfig, cfg.Recipients, subject, template, data); err != nil {
		log.Printf("[ERROR] Falha ao enviar e-mail (%s): %v\n", serviceName, err)
	}
}

func selectTemplate(status string) string {
	switch status {
	case "Stopped":
		return "email/templates/service_stopped.html"
	case "Started":
		return "email/templates/service_started.html"
	case "StartFailed":
		return "email/templates/service_failed.html"
	case "ResourceAlert":
		return "email/templates/alerta_recursos.html"
	case "MaxRetries":
		return "email/templates/max_retries.html"
	default:
		return "email/templates/generic_alert.html"
	}
}
