package logs

import (
	"os"
	"path/filepath"
)

func OpenLogFile() (*os.File, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	logDir := filepath.Join(baseDir, "Logs")
	logPath := filepath.Join(logDir, "logs.txt")

	// Cria o diretório Logs se não existir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Abre/cria o arquivo logs.txt
	file, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}
