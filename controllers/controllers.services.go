package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"ANFeGuard/database"
	"ANFeGuard/winservice"
)

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

func HandleServices(w http.ResponseWriter, r *http.Request) {
	type Service struct {
		ID     int               `json:"id"`
		Nome   string            `json:"nome"`
		Status winservice.Status `json:"status"`
	}

	rows, err := database.DB.Query(`SELECT id, nome, status FROM servicos ORDER BY nome ASC`)
	if err != nil {
		log.Println("Erro ao listar serviços:", err)
		http.Error(w, "Erro ao listar serviços", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var s Service
		var statusDB int

		if err := rows.Scan(&s.ID, &s.Nome, &statusDB); err != nil {
			http.Error(w, "Erro ao ler serviços", http.StatusInternalServerError)
			return
		}

		winStatus, err := winservice.GetStatus(s.Nome)
		if err == nil {
			s.Status = winStatus
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

func DeletarServico(w http.ResponseWriter, r *http.Request) {
	log.Println("DELETE chamado em:", r.URL.Path)

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID não informado", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`DELETE FROM servicos WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Erro ao deletar serviço: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func EditarServico(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/servico/")
	if id == "" {
		http.Error(w, "ID não informado", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`UPDATE servicos SET status = 1 - status WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Erro ao editar serviço: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
