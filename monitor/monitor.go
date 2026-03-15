package monitor

import (
	"ANFeGuard/database"
	"ANFeGuard/email"
	"ANFeGuard/logs"
	"ANFeGuard/sysinfo"
	"ANFeGuard/winservice"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
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
	// TimeOut		int8
}

type ServerData struct {
	CPU    float64
	Memory float64
	Disk   float64
}

// Para diminuur as consultas de recursos
var serverData ServerData
var serverDataMutex sync.RWMutex

// Mediana de consumo do servidor
var cpuSamples []float64
var memSamples []float64
var samplesMutex sync.Mutex

const maxSamples = 10

// Usado para controlar qual ciclo e go rotine gerou o log
var cycleID uint64 = 0
var goroutineCounter uint64

// Controla tentativas por serviço
var retryCount = map[string]int{}

// Controla se o e-mail por
var reachMaxRetries = map[string]bool{}

// Adicionada para travar a escrita no retryCount
var retryMutex sync.Mutex

// Para evitar execução simultanea na função Start
var monitorRunning bool
var monitorMutex sync.Mutex

// =====================================================
// INICIO DO MONITORAMENTO
// =====================================================

func Start(cfg MonitorConfig) {
	startLogWorker()
	logs.Info("ANFeGuard Monitor iniciado | Intervalo: %s\n", cfg.Period)
	logs.Info("Goroutines em execução: %d", runtime.NumGoroutine())

	ticker := time.NewTicker(cfg.Period)
	defer ticker.Stop()

	/*
	   15/03/26 alterado para criar uma go routine para cada função e evitar que o processo pare por causa de panic
	   Também controla se ainda existe uma função em execução ou não antes de criar novas go routine
	*/

	for range ticker.C {
		cycle := atomic.AddUint64(&cycleID, 1)

		monitorMutex.Lock()
		if monitorRunning {
			monitorMutex.Unlock()
			logs.Warn("[cycle:%d] Monitor ainda está executando, pulando ciclo", cycle)
			continue
		}

		monitorRunning = true
		monitorMutex.Unlock()

		go func(cycle uint64) {

			defer func() {
				monitorMutex.Lock()
				monitorRunning = false
				monitorMutex.Unlock()
			}()
			gid := nextGID()
			safeExecute("monitorSystem", func() {
				monitorSystem(cfg, cycle, gid)
			})

			safeExecute("monitorServices", func() {
				monitorServices(cfg, cycle, gid)
			})

		}(cycle)

	}
}

// =====================================================
// MONITORAMENTO DE SISTEMA (CPU / RAM / DISCO)
// =====================================================
func monitorSystem(cfg MonitorConfig, cycle uint64, gid uint64) {
	logs.Info("[cycle:%d gid:%d] Goroutines em execução: %d",
		cycle, gid, runtime.NumGoroutine())

	defer logs.TrackWithId("Buscar recursos do servidor", cycle, gid)()
	info, err := sysinfo.GetSystemInfo(cfg.DiskPath)
	if err != nil {
		logs.Error("Falha ao coletar informações de sistema: %v\n", err)
		return
	}

	samplesMutex.Lock()

	addSample(&cpuSamples, info.CPUPercent)
	addSample(&memSamples, info.MemoryPercent)

	cpuMedian := median(cpuSamples)
	memMedian := median(memSamples)

	samplesMutex.Unlock()

	updateServerData(
		cpuMedian,
		memMedian,
		info.DiskUsedPercent,
	)
	data := getServerData()

	/*
		31/12/25 - removi para não ficar poluindo o log, agora só gera log quando houver um alto consumo
		 now := time.Now().Format("02/01/2006 15:04:05")
		 logs.Info("%s | CPU: %.1f%% | RAM: %.1f%% | Disco(%s): %.1f%%", now,
		  info.CPUPercent, info.MemoryPercent, cfg.DiskPath, info.DiskUsedPercent)
	*/

	// Verifica limites
	if data.CPU > cfg.CPULimit || data.Memory > cfg.MemLimit {
		logs.Warn("Limites de recursos excedidos (mediana CPU/RAM)")
		sendServiceEmail(cfg, "", "ResourceAlert", "Alerta ANFeGuard — Uso elevado de recursos")
	}
}

// =====================================================
// MONITORAMENTO DE SERVIÇOS
// =====================================================
func monitorServices(cfg MonitorConfig, cycle uint64, gid uint64) {
	logs.Info("[cycle:%d gid:%d] Goroutines em execução: %d",
		cycle, gid, runtime.NumGoroutine())

	defer logs.TrackWithId("Verificando serviços", cycle, gid)()
	services := database.GetServices()
	sysInfo := getServerData()

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
func retryServiceStart(cfg MonitorConfig, svc string, status winservice.Status, sysInfo *ServerData) {

	retryMutex.Lock()
	retryCount[svc]++
	retry := retryCount[svc]
	maxReached := reachMaxRetries[svc]
	retryMutex.Unlock()

	// Excedeu tentativas
	if retry > cfg.MaxRetries && !maxReached {
		logs.Error("Serviço '%s' atingiu o máximo de tentativas (%d)\n", svc, cfg.MaxRetries)
		sendServiceEmail(cfg, svc, "MaxRetries", "ANFeGuard — Máximo de tentativas atingido")

		retryMutex.Lock()
		reachMaxRetries[svc] = true
		retryMutex.Unlock()

		return
	}

	logs.Warn("Serviço '%s' está parado. Tentativa %d/%d\n",
		svc, retry, cfg.MaxRetries)

	logServiceError(svc, string(status), "Serviço parado", sysInfo)

	sendServiceEmail(cfg, svc, "Stopped", "Serviço parado — Tentando iniciar...")

	// Tentar iniciar
	if err := winservice.Start(svc); err != nil {
		msg := fmt.Sprintf("Falha ao iniciar: %v", err)

		logs.Error("Falha ao iniciar %s: %v", svc, err)

		sendServiceEmail(cfg, svc, "StartFailed", "Falha ao iniciar serviço!")
		logServiceError(svc, string(status), msg, sysInfo)

		return
	}

	// Sucesso — reseta tentativas
	resetRetries(svc)

	logs.Success("Serviço '%s' iniciado com sucesso.\n", svc)
	sendServiceEmail(cfg, svc, "Started", "Serviço iniciado com sucesso!")
}

// =====================================================
// RESET DE RETENTATIVAS
// =====================================================
func resetRetries(service string) {

	retryMutex.Lock()
	defer retryMutex.Unlock()
	if retryCount[service] > 0 {
		logs.Info("Resetando tentativas do serviço '%s'\n", service)
	}
	retryCount[service] = 0
	reachMaxRetries[service] = false
}

// =====================================================
// LOG DE ERRO CENTRALIZADO
// =====================================================
func logServiceError(svc, status, msg string, info *ServerData) {
	logs.Error("Serviço '%s' | Status: %s | %s\n", svc, status, msg)
	//go database.LogServiceError(svc, status, msg, info.Memory)

	//Foi criada uma fila para evitar que muitas go routines sejam criadas

	logQueue <- serviceLog{
		service: svc,
		status:  status,
		message: msg,
		memory:  info.Memory,
	}
}

// =====================================================
// ENVIO DE EMAIL CENTRALIZADO
// =====================================================
func sendServiceEmail(cfg MonitorConfig, serviceName, status string, subject string) {

	dataServer := getServerData()

	data := email.EmailAlertData{
		Service:  serviceName,
		CPU:      fmt.Sprintf("%.2f%%", dataServer.CPU),
		Memory:   fmt.Sprintf("%.2f%%", dataServer.Memory),
		Disk:     fmt.Sprintf("%.2f%%", dataServer.Disk),
		DiskPath: cfg.DiskPath,
		Time:     time.Now().Format("02/01/2006 15:04:05"),
	}

	template := selectTemplate(status)

	// if err := email.SendEmail(cfg.EmailConfig, cfg.Recipients, subject, template, data); err != nil {
	// 	logs.Error("Falha ao enviar e-mail (%s): %v\n", serviceName, err)
	// }

	go func(service string) {

		if err := email.SendEmail(cfg.EmailConfig, cfg.Recipients, subject, template, data); err != nil {
			logs.Error("Falha ao enviar e-mail (%s): %v", service, err)
		}

	}(serviceName)
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

// =====================================================
// Proteção contra panic
// =====================================================

func safeExecute(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logs.Error("Erro na execução da rotina: %s: %v", name, r)

		}
	}()

	fn()
}

func nextGID() uint64 {
	return atomic.AddUint64(&goroutineCounter, 1)
}

func updateServerData(cpu, mem, disk float64) {
	serverDataMutex.Lock()
	serverData.CPU = cpu
	serverData.Memory = mem
	serverData.Disk = disk
	serverDataMutex.Unlock()
}

func getServerData() ServerData {
	serverDataMutex.RLock()
	data := serverData
	serverDataMutex.RUnlock()
	return data
}

func addSample(arr *[]float64, value float64) {

	if len(*arr) >= maxSamples {
		*arr = (*arr)[1:]
	}

	*arr = append(*arr, value)
}

func median(values []float64) float64 {

	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)

	sort.Float64s(sorted)

	middle := len(sorted) / 2

	if len(sorted)%2 == 0 {
		return (sorted[middle-1] + sorted[middle]) / 2
	}

	return sorted[middle]
}
