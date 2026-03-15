package database

import (
	"ANFeGuard/logs"
	"context"
	"time"
)

func LogServiceError(serviceName, status, message string, memory float64) {
	timeOut := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	_, err := DB.ExecContext(ctx,
		`INSERT INTO service_logs (service_name, status, message, memory_percent)
         VALUES ($1, $2, $3, $4)`,
		serviceName, status, message, memory,
	)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {

			logs.TimeOut("Banco demorou para registrar log do serviço '%s'", serviceName)
			return
		}

		logs.Error("Falha ao registrar log do serviço '%s': %v\n", serviceName, err)
	}
}
