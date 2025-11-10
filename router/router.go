package router

import (
	"ANFeGuard/controllers"
	"ANFeGuard/database"
	"ANFeGuard/sysinfo"
	"ANFeGuard/winservice"
	"encoding/json"

	"net/http"
)

func SetupRoutes(mux *http.ServeMux) {

	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/", fs)

	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/metrics", handleMetrics)

	mux.HandleFunc("/api/servicos", handleServices)
	mux.HandleFunc("POST /api/servico", CriarServico)

	mux.HandleFunc("/api/portas", controllers.ListarPortas)
	mux.HandleFunc("POST /api/porta", controllers.CriarPorta)
	mux.HandleFunc("DELETE /api/porta/{id}", controllers.DeletarPorta)

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

func handleServices(w http.ResponseWriter, r *http.Request) {
	type Service struct {
		Nome   string            `json:"nome"`
		Status winservice.Status `json:"status"`
	}

	rows, err := database.DB.Query(`SELECT nome FROM servicos ORDER BY nome ASC`)
	if err != nil {
		http.Error(w, "Erro ao listar serviços", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.Nome); err != nil {
			return
		}

		statusAtual, err := winservice.GetStatus(s.Nome)
		if err == nil {
			s.Status = statusAtual
		}

		services = append(services, s)

	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erro ao processar resultados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func CriarServico(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Lê os dados do form (vindo do HTMX)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao ler formulário: "+err.Error(), http.StatusBadRequest)
		return
	}

	nome := r.FormValue("servicename")
	if nome == "" {
		http.Error(w, "O campo 'servicename' é obrigatório", http.StatusBadRequest)
		return
	}

	// Insere no banco
	_, err := database.DB.Exec(`INSERT INTO servicos (nome) VALUES ($1)`, nome)
	if err != nil {
		http.Error(w, "Erro ao inserir serviço: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna algum conteúdo para HTMX (pode ser vazio, mas não JSON inválido)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(nome)) // HTMX aceita texto simples
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
