package database

import (
	"ANFeGuard/logs"
	//"ANFeGuard/models"
)

type Service struct {
	Name string
	Host string
}

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

func GetServicesApi() []Service {
	rows, err := DB.Query(`
	SELECT
	nome, urlApi
	FROM servicos 
	WHERE ativo = 1
	AND urlApi IS NOT NULL 
	AND urlApi <> ''
	ORDER BY nome ASC`)
	if err != nil {
		logs.Error("Falha ao listar serviços: %v", err)
		return []Service{}
	}
	defer rows.Close()

	var services []Service

	for rows.Next() {
		var svc Service

		if err := rows.Scan(&svc.Name, &svc.Host); err != nil {
			logs.Error("Falha ao ler serviço: %v", err)
			continue
		}

		services = append(services, svc)
	}

	if err := rows.Err(); err != nil {
		logs.Error("Erro ao processar resultados: %v", err)
	}

	return services
}
