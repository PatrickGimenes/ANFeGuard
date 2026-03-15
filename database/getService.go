package database

import "ANFeGuard/logs"

// GetServices retorna todos os nomes de serviços cadastrados no banco
func GetServices() []string {
	rows, err := DB.Query(`SELECT nome FROM servicos WHERE ativo = 1 ORDER BY nome ASC`)
	if err != nil {
		logs.Error("Falha ao listar serviços:", err)
		return []string{}
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var nome string
		if err := rows.Scan(&nome); err != nil {
			logs.Error("Falha ao ler serviço:", err)
			continue
		}
		services = append(services, nome)
	}

	if err := rows.Err(); err != nil {
		logs.Error("Erro ao processar resultados:", err)
	}

	return services
}
