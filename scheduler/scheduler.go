package scheduler

import (
	"ANFeGuard/database"
	"ANFeGuard/logs"
	"ANFeGuard/winservice"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var restartControl = map[string]string{}   // serviço -> data (yyyy-mm-dd)
var retrySchedule = map[string]time.Time{} // serviço -> próxima tentativa
var restartMutex sync.Mutex

func StartScheduler() {

	nextStr := os.Getenv("NEXT_TRY")

	next, err := strconv.Atoi(nextStr)
	if err != nil || next <= 0 {
		logs.Warn("NEXT_TRY inválido (%s), usando padrão 30 minutos", nextStr)
		next = 30
	}

	ticker := time.NewTicker(time.Duration(next) * time.Minute)
	defer ticker.Stop()

	logs.Info("Scheduler de restart iniciado")

	lastDay := ""
	for range ticker.C {

		currentDay := time.Now().Format("2006-01-02")

		if currentDay != lastDay {
			resetDailyControl()
			lastDay = currentDay
		}

		// Só roda na janela 22h até às 6h
		if !isRestartWindow() {
			continue
		}

		runCycle()
	}
}

func runCycle() {
	defer logs.Track("Scheduler - cliclo principal")()
	services := database.GetServicesApi()

	allDone := true

	for _, svc := range services {

		if alreadyRestartedToday(svc.Name) {
			logs.Info("Este serviço %s já foi reiniciado hoje!", svc.Name)
			continue
		}

		allDone = false

		if !canRetryNow(svc.Name) {
			continue
		}

		blocked, err := canRestartFromAPI(svc.Host)
		if err != nil {
			logs.Error("Erro API %s: %v", svc.Name, err)
			continue
		}

		if !blocked {
			logs.Info("Serviço '%s' tem rotina em execução. Nova tentativa em 30min", svc.Name)
			scheduleRetry(svc.Name)
			continue
		}

		logs.Warn("Reiniciando serviço '%s'", svc.Name)

		winservice.RestartService(svc.Name)

		markRestartedToday(svc.Name)

		logs.Success("Serviço '%s' reiniciado", svc.Name)
	}

	if allDone {
		logs.Info("Todos os serviços já foram reiniciados hoje")
	}
}

func isRestartWindow() bool {
	now := time.Now().Hour()

	start, err := strconv.Atoi(os.Getenv("START"))
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return false
	}
	end, err := strconv.Atoi(os.Getenv("END"))
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return false
	}

	return now >= start || now < end
}

func alreadyRestartedToday(svc string) bool {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	today := time.Now().Format("2006-01-02")
	return restartControl[svc] == today
}

func markRestartedToday(svc string) {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	today := time.Now().Format("2006-01-02")
	restartControl[svc] = today
}

func canRetryNow(svc string) bool {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	next, exists := retrySchedule[svc]
	if !exists {
		return true
	}

	return time.Now().After(next)
}

func scheduleRetry(svc string) {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	retrySchedule[svc] = time.Now().Add(30 * time.Minute)
}

func canRestartFromAPI(svc string) (bool, error) {
	token := os.Getenv("TOKEN")

	req, err := http.NewRequest("POST", svc, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status inválido: %d", resp.StatusCode)
	}

	// lê tudo primeiro
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	body := string(bodyBytes)

	// DEBUG
	//logs.Info("Resposta bruta: %s", body)

	// remove BOM (isso resolve o erro 'ï')
	body = strings.TrimPrefix(body, "\uFEFF")

	// JSON externo
	var outer struct {
		Result string `json:"getRotinasCriticasEmExecucaoResult"`
	}

	if err := json.Unmarshal([]byte(body), &outer); err != nil {
		return false, fmt.Errorf("erro outer: %v | body: %s", err, body)
	}

	// JSON interno
	var inner struct {
		Success bool   `json:"sucess"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(outer.Result), &inner); err != nil {
		return false, fmt.Errorf("erro inner: %v | inner: %s", err, outer.Result)
	}

	logs.Info("API retorno: success=%v message=%s", inner.Success, inner.Message)

	// true → pode reiniciar
	// false → NÃO pode reiniciar
	return inner.Success, nil
}

func resetDailyControl() {
	restartMutex.Lock()
	defer restartMutex.Unlock()

	restartControl = map[string]string{}
}
