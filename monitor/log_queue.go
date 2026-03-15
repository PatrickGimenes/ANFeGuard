package monitor

import "ANFeGuard/database"

type serviceLog struct {
	service string
	status  string
	message string
	memory  float64
}

var logQueue = make(chan serviceLog, 100)

func startLogWorker() {

	go func() {

		for logItem := range logQueue {

			database.LogServiceError(
				logItem.service,
				logItem.status,
				logItem.message,
				logItem.memory,
			)

		}

	}()

}