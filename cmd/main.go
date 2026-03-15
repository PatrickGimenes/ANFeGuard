package main

import (
	"ANFeGuard/database"
	"ANFeGuard/email"
	"ANFeGuard/logs"
	"ANFeGuard/monitor"
	"ANFeGuard/router"
	"ANFeGuard/version"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	//"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	//define a informações que serão exibidas no log: 2026/03/15 20:42:11.234567 monitor.go:45: [INFO] isso é um exemplo
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	logs.Info("Versão atual:  %s", version.Version)

	logFile, err := logs.OpenLogFile()
	if err != nil {
		logs.Critical("Erro ao abrir arquivo de log: %v", err)
	}

	// Tela + arquivo
	mw := io.MultiWriter(os.Stdout, logFile)

	// Define saída global para o logger
	log.SetOutput(mw)

	godotenv.Load()

	// Conecta ao banco
	if err := database.Conectar(); err != nil {
		logs.Critical("Falha ao iniciar banco: %v", err)
	}

	mux := http.NewServeMux()
	router.SetupRoutes(mux)

	period, err := strconv.Atoi(os.Getenv("PERIOD"))
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return
	}

	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return
	}

	max, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return
	}

	// timeOut, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	// if err != nil {
	// 	logs.Error("Erro ao converter:", err)
	// 	return
	// }

	limit, err := strconv.ParseFloat(os.Getenv("THRESHOLD_WARNING"), 64) // 64 é a precisão (float64)
	if err != nil {
		logs.Error("Erro ao converter:", err)
		return
	}

	emails := os.Getenv("NOTIFY_EMAILS")
	recipients := strings.Split(emails, ",")

	cfg := monitor.MonitorConfig{

		Period: time.Duration(period) * time.Second,
		EmailConfig: email.SMTPConfig{
			Host:     os.Getenv("EMAIL_HOST"),
			Port:     port,
			User:     os.Getenv("EMAIL_USER"),
			Password: os.Getenv("EMAIL_PASS"),
			From:     os.Getenv("EMAIL_USER"),
		},
		Recipients: recipients,
		MaxRetries: max,
		CPULimit:   limit,
		MemLimit:   limit,
		DiskPath:   os.Getenv("DISK"),
		// TimeOut: int8(timeOut),
	}

	go monitor.Start(cfg)
	API_port := os.Getenv("API_PORT")
	if API_port == "" {
		API_port = "30000" // porta padrão
	}
	addr := ":" + API_port // forma correta para ListenAndServe
	logs.Info("Servidor rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logs.Critical("Erro ao iniciar servidor: %v", err)
	}
}
