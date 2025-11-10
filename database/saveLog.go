package database

import "log"

func LogServiceError(serviceName, status, message string, memory float64) {
	_, err := DB.Exec(
		`INSERT INTO service_logs (service_name, status, message, memory_percent)
         VALUES ($1, $2, $3, $4)`,
		serviceName, status, message, memory,
	)
	if err != nil {
		log.Printf("[ERRO] Falha ao registrar log do serviço '%s': %v\n", serviceName, err)
	}
}
