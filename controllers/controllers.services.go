package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"ANFeGuard/database"
	"ANFeGuard/logs"
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

	chave := r.FormValue("chave")
	if chave == "" {
		http.Error(w, "O campo 'chave' é obrigatório", http.StatusBadRequest)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "O campo 'url' é obrigatório", http.StatusBadRequest)
		return
	}

	// Insere no banco
	_, err := database.DB.Exec(`INSERT INTO servicos (nome, displayname, chave, urlApi) VALUES ($1, $2, $3, $4)`, ServiceName, displayName, chave, url)
	if err != nil {
		http.Error(w, "Erro ao inserir serviço: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna algum conteúdo para HTMX (pode ser vazio, mas não JSON inválido)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ServiceName)) // HTMX aceita texto simples
}

func HandleServices(w http.ResponseWriter, r *http.Request) {

	type ErrorResponse struct {
		Error string `json:"error"`
	}

	//mover para models/Service.go
	type Service struct {
		ID          int               `json:"id"`
		Nome        string            `json:"nome"`
		DisplayName string            `json:"displayname"`
		Ativo       bool              `json:"ativo"`
		Status      winservice.Status `json:"status"`
		Chave       string            `json:"chave"`
		URL         string            `json:"url"`
	}

	rows, err := database.DB.Query(`SELECT id, nome, displayname, ativo, chave, urlApi FROM servicos ORDER BY nome ASC`)
	if err != nil {
		logs.Error("Erro ao listar serviços:", err)
		//http.Error(w, "Erro ao listar serviços", http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Erro ao ler serviços",
		})
		return
	}
	defer rows.Close()

	var services []Service
	var url sql.NullString

	for rows.Next() {
		var s Service

		if err := rows.Scan(&s.ID, &s.Nome, &s.DisplayName, &s.Ativo, &s.Chave, &url); err != nil {
			//http.Error(w, "Erro ao ler serviços", http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error: "Erro ao ler serviços",
			})
			return
		}
		if url.Valid {
			s.URL = url.String
		}
		if s.Ativo {
			winStatus, err := winservice.GetStatus(s.Nome)
			if err != nil {
				logs.Error("Erro ao obter status do serviço %s: %v", s.Nome, err)
			} else {
				s.Status = winStatus
			}

		} else {
			s.Status = winservice.StatusUnknown
		}

		services = append(services, s)
	}

	if err := rows.Err(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Erro ao ler serviços",
		})
		//http.Error(w, "Erro ao processar resultados", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)

}

func DeletarServico(w http.ResponseWriter, r *http.Request) {
	logs.Router("DELETE chamado em:", r.URL.Path)

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

func RestartService(w http.ResponseWriter, r *http.Request) {
	defer logs.Track("POST reiniciar serviço")()
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	type RestartRequest struct {
		Chave string `json:"chave"`
	}

	type Service struct {
		Nome string `json:"nome"`
	}

	var req RestartRequest

	// Decodifica o JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	logs.Router("%s | %s", r.URL.Path, req.Chave)

	rows, err := database.DB.Query(
		`SELECT nome FROM servicos WHERE chave = $1`,
		req.Chave,
	)
	if err != nil {
		logs.Error("Erro ao buscar serviço:", err)
		http.Error(w, "Erro ao buscar serviço", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var encontrado bool

	for rows.Next() {
		encontrado = true

		var s Service

		if err := rows.Scan(&s.Nome); err != nil {
			http.Error(w, "Erro ao ler serviço", http.StatusInternalServerError)
			return
		}

		err := winservice.RestartService(s.Nome)
		if err != nil {
			logs.Error("Erro ao reiniciar serviço %s: %v", s.Nome, err)
			http.Error(w, "Erro ao reiniciar serviço", http.StatusInternalServerError)
			return
		}
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Erro ao processar resultados", http.StatusInternalServerError)
		return
	}

	if !encontrado {
		http.Error(w, "Serviço não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Serviço reiniciado com sucesso"))
}
