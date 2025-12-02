package router

import (
	"ANFeGuard/controllers"
	"ANFeGuard/database"
	"ANFeGuard/sysinfo"
	"encoding/json"

	"net/http"
)

func SetupRoutes(mux *http.ServeMux) {

	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/metrics", handleMetrics)

	mux.HandleFunc("GET /api/servicos", controllers.HandleServices)
	mux.HandleFunc("POST /api/servico", controllers.CriarServico)


	mux.HandleFunc("/api/portas", controllers.ListarPortas)
	mux.HandleFunc("/api/porta", controllers.CriarPorta)
	mux.HandleFunc("/api/porta/{id}", controllers.DeletarPorta)

	mux.HandleFunc("/api/logs", HandleLogs)

}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	info, err := sysinfo.GetSystemInfo("C:\\")

	if err != nil {
		http.Error(w, "Failed to get system info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encode := json.NewEncoder(w)
	encode.SetIndent("", " ")
	_ = encode.Encode(info)
}

type ServiceLog struct {
	ID            int     `json:"id"`
	ServiceName   string  `json:"service_name"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	MemoryPercent float64 `json:"memory_percent"`
	CreatedAt     string  `json:"created_at"`
}

func HandleLogs(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT id, service_name, status, message, memory_percent, created_at
		FROM service_logs
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		http.Error(w, "Erro ao listar logs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []ServiceLog
	for rows.Next() {
		var l ServiceLog
		if err := rows.Scan(&l.ID, &l.ServiceName, &l.Status, &l.Message, &l.MemoryPercent, &l.CreatedAt); err != nil {
			http.Error(w, "Erro ao ler log: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
