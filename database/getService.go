package database

import "log"

// GetServices retorna todos os nomes de serviços cadastrados no banco
func GetServices() []string {
	rows, err := DB.Query(`SELECT nome FROM servicos ORDER BY nome ASC`)
	if err != nil {
		log.Println("[ERRO] Falha ao listar serviços:", err)
		return []string{}
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var nome string
		if err := rows.Scan(&nome); err != nil {
			log.Println("[ERRO] Falha ao ler serviço:", err)
			continue
		}
		services = append(services, nome)
	}

	if err := rows.Err(); err != nil {
		log.Println("[ERRO] Erro ao processar resultados:", err)
	}

	return services
}