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

	ServiceName := r.FormValue("servicename")
	if ServiceName == "" {
		http.Error(w, "O campo 'servicename' é obrigatório", http.StatusBadRequest)
		return
	}
	displayName := r.FormValue("displayname")
	if displayName == "" {
		http.Error(w, "O campo 'displayname' é obrigatório", http.StatusBadRequest)
		return
	}

	// Insere no banco
	_, err := database.DB.Exec(`INSERT INTO servicos (nome, displayname) VALUES ($1, $2)`, ServiceName, displayName)
	if err != nil {
		http.Error(w, "Erro ao inserir serviço: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna algum conteúdo para HTMX (pode ser vazio, mas não JSON inválido)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ServiceName)) // HTMX aceita texto simples
}

func HandleServices(w http.ResponseWriter, r *http.Request) {

	//mover para models/Service.go
	type Service struct {
		ID          int               `json:"id"`
		Nome        string            `json:"nome"`
		DisplayName string            `json:"displayname"`
		Ativo       bool              `json:"ativo"`
		Status      winservice.Status `json:"status"`
	}

	rows, err := database.DB.Query(`SELECT id, nome, displayname, ativo FROM servicos ORDER BY nome ASC`)
	if err != nil {
		log.Println("Erro ao listar serviços:", err)
		http.Error(w, "Erro ao listar serviços", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var s Service

		if err := rows.Scan(&s.ID, &s.Nome, &s.DisplayName, &s.Ativo); err != nil {
			http.Error(w, "Erro ao ler serviços", http.StatusInternalServerError)
			return
		}

		if s.Ativo {
			winStatus, err := winservice.GetStatus(s.Nome)
			if err != nil {
				log.Printf("[Controller] Erro ao obter status do serviço %s: %v", s.Nome, err)
			} else {
				s.Status = winStatus
			}

		} else {
			s.Status = winservice.StatusUnknown
		}

		services = append(services, s)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erro ao processar resultados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
