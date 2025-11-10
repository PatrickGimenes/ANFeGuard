package main

import (
	"ANFeGuard/database"
	"ANFeGuard/email"
	"ANFeGuard/monitor"
	"ANFeGuard/router"
	"os"
	"strconv"
	"time"

	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()

	// Conecta ao banco
	database.Conectar()

	mux := http.NewServeMux()
	router.SetupRoutes(mux)
	fmt.Println("Servidor rodando em :8080 e monitorando recursos...")

	period, err := strconv.Atoi(os.Getenv("PERIOD"))
	if err != nil {
		fmt.Println("Erro ao converter:", err)
		return
	}

	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		fmt.Println("Erro ao converter:", err)
		return
	}

	max, err := strconv.Atoi(os.Getenv("MAX_RETRIES"))
	if err != nil {
		fmt.Println("Erro ao converter:", err)
		return
	}

	limit, err := strconv.ParseFloat(os.Getenv("THRESHOLD_WARNING"), 64) // 64 é a precisão (float64)
	if err != nil {
		fmt.Println("Erro ao converter:", err)
		return
	}
	cfg := monitor.MonitorConfig{

		Period:   time.Duration(period) * time.Second,
		Services: database.GetServices(),
		EmailConfig: email.SMTPConfig{
			Host:     os.Getenv("EMAIL_HOST"),
			Port:     port,
			User:     os.Getenv("EMAIL_USER"),
			Password: os.Getenv("EMAIL_PASS"),
			From:     os.Getenv("EMAIL_USER"),
		},
		Recipients: []string{os.Getenv("NOTIFY_EMAIL")},
		MaxRetries: max,
		CPULimit:   limit,
		MemLimit:   limit,
		DiskPath:   os.Getenv("DISK"),
	}

	go monitor.Start(cfg)
	API_port := os.Getenv("API_PORT")
	if API_port == "" {
		API_port = "8080" // porta padrão
	}
	addr := ":" + API_port // forma correta para ListenAndServe
	log.Printf("Servidor rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
