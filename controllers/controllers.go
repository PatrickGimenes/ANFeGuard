package controllers

import (
	"fmt"
	"html/template"
	"net/http"

	"ANFeGuard/database"
)

type Porta struct {
	ID       int
	Nome     string
	PortaWS  int
	PortaAPI int
	Ambiente string
}

// POST /api/porta
func CriarPorta(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	s := Porta{
		Nome:     r.FormValue("nome"),
		PortaWS:  atoi(r.FormValue("porta_ws")),
		PortaAPI: atoi(r.FormValue("porta_api")),
		Ambiente: r.FormValue("ambiente"),
	}

	_, err := database.DB.Exec(`
		INSERT INTO portas (nome, porta_ws, porta_api, ambiente)
		VALUES ($1, $2, $3, $4)
	`, s.Nome, s.PortaWS, s.PortaAPI, s.Ambiente)

	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao salvar serviço: %v", err), http.StatusInternalServerError)
		return
	}

	// Retorna a nova linha da tabela em HTML (para o htmx inserir)
	tmpl := `<tr>
		<td>{{.Nome}}</td>
		<td>{{.PortaWS}}</td>
		<td>{{.PortaAPI}}</td>
		<td>{{.Ambiente}}</td>
	</tr>`

	t := template.Must(template.New("row").Parse(tmpl))
	t.Execute(w, s)
}

// GET /api/portas
func ListarPortas(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`SELECT id, nome, porta_ws, porta_api, ambiente FROM portas ORDER BY porta_ws`)
	if err != nil {
		http.Error(w, "Erro ao listar serviços", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tmpl := template.Must(template.New("linha").Parse(`
	<tr id="porta-{{.ID}}">
		<td>{{.Nome}}</td>
		<td>{{.PortaWS}}</td>
		<td>{{.PortaAPI}}</td>
		<td>{{.Ambiente}}</td>
		<td>
			<button 
			hx-delete="/api/porta/{{.ID}}" 
			hx-target="#porta-{{.ID}}" 
			hx-swap="outerHTML"
			hx-on="htmx:afterOnLoad: htmx.trigger('#tabela-portas tbody', 'refresh')"
			class="btn-delete">
			🗑 Excluir
			</button>
		</td>
	</tr>`))

	var count int
	for rows.Next() {
		var s Porta
		if err := rows.Scan(&s.ID, &s.Nome, &s.PortaWS, &s.PortaAPI, &s.Ambiente); err != nil {
			http.Error(w, "Erro ao ler dados", http.StatusInternalServerError)
			return
		}
		count++
		tmpl.Execute(w, s)
	}

	// Caso não tenha nenhum serviço
	if count == 0 {
		fmt.Fprint(w, `<tr><td colspan="5" class="empty">Nenhuma porta cadastrada.</td></tr>`)
	}
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func DeletarPorta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec("DELETE FROM portas WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Erro ao excluir porta", http.StatusInternalServerError)
		return
	}

	// Retorna vazio para remover a linha
	w.WriteHeader(http.StatusNoContent)
}
