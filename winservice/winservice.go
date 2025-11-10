package winservice

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc/mgr"
)

type Status string

const (
	StatusRunning  Status = "Running"
	StatusStopped  Status = "Stopped"
	StatusStarting Status = "Starting"
	StatusStopping Status = "Stopping"
	StatusUnknown  Status = "Unknown"
)

// GetStatus retorna o estado atual do serviço
func GetStatus(serviceName string) (Status, error) {
	m, err := mgr.Connect()
	if err != nil {
		return StatusUnknown, fmt.Errorf("erro ao abrir gerenciador de serviços: %w", err)
	}
	defer m.Disconnect()

	srv, err := m.OpenService(serviceName)
	if err != nil {
		return StatusUnknown, fmt.Errorf("serviço '%s' não encontrado: %w", serviceName, err)
	}
	defer srv.Close()

	status, err := srv.Query()
	if err != nil {
		return StatusUnknown, fmt.Errorf("erro ao consultar status do serviço: %w", err)
	}

	switch status.State {
	case 1:
		return StatusStopped, nil
	case 2:
		return StatusStarting, nil
	case 3:
		return StatusStopping, nil
	case 4:
		return StatusRunning, nil
	default:
		return StatusUnknown, nil
	}
}

// Start inicia o serviço e espera até que esteja em execução
func Start(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar ao gerenciador de serviços: %w", err)
	}
	defer m.Disconnect()

	srv, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("erro ao abrir serviço '%s': %w", serviceName, err)
	}
	defer srv.Close()

	err = srv.Start()
	if err != nil {
		return fmt.Errorf("erro ao iniciar serviço '%s': %w", serviceName, err)
	}

	// Espera o serviço entrar em estado de execução
	for i := 0; i < 20; i++ {
		time.Sleep(500 * time.Millisecond)
		status, _ := srv.Query()
		if status.State == 4 { // SERVICE_RUNNING
			return nil
		}
	}

	return fmt.Errorf("tempo limite atingido ao iniciar o serviço '%s'", serviceName)
}

// Restart reinicia o serviço
func Restart(serviceName string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("erro ao conectar ao gerenciador de serviços: %w", err)
	}
	defer m.Disconnect()

	srv, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("erro ao abrir serviço '%s': %w", serviceName, err)
	}
	defer srv.Close()

	status, err := srv.Control(1) // SERVICE_CONTROL_STOP
	if err != nil {
		return fmt.Errorf("erro ao enviar comando de parada: %w", err)
	}

	// Espera até parar
	for status.State != 1 {
		time.Sleep(500 * time.Millisecond)
		status, _ = srv.Query()
		if status.State == 1 {
			break
		}
	}

	err = srv.Start()
	if err != nil {
		return fmt.Errorf("erro ao reiniciar serviço '%s': %w", serviceName, err)
	}

	return nil
}
