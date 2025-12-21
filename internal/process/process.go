package process

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// GetPidByName busca el PID de un proceso a partir de su nombre (por ejemplo, "mariadbd") usando pgrep (o ps si pgrep no está disponible) y devuelve el PID (o un error si no se encuentra).
func GetPidByName(processName string) (int, error) {
	// Intentar usar pgrep (más rápido y común en Linux)
	cmd := exec.Command("pgrep", processName)
	output, err := cmd.Output()
	if err == nil {
		// pgrep devuelve el PID (o varios, uno por línea) en su salida estándar.
		// Aquí asumimos que solo se devuelve un PID (el primero si hay varios).
		pidStr := strings.TrimSpace(string(output))
		lines := strings.Split(pidStr, "\n")
		if len(lines) == 0 || lines[0] == "" {
			return 0, fmt.Errorf("no process found with name '%s'", processName)
		}
		pid, err := strconv.Atoi(lines[0])
		if err != nil {
			return 0, fmt.Errorf("error parsing pgrep output ('%v'): %v", lines[0], err)
		}
		return pid, nil
	}

	// Si pgrep falla (por ejemplo, no está instalado o no se encuentra el proceso), intentar con "ps" (más lento pero más común).
	// Ejemplo: "ps aux | grep [m]ariadbd" (usando "[" para evitar que grep se capture a sí mismo).
	cmd = exec.Command("sh", "-c", fmt.Sprintf("ps aux | grep [%c]%s", processName[0], processName[1:]))
	output, err = cmd.Output()
	if err != nil {
		// Si "ps" también falla, devolver un error.
		return 0, fmt.Errorf("error running ps (or pgrep) for '%s': %v", processName, err)
	}
	// "ps" devuelve líneas (una por proceso) con el PID en la segunda columna (índice 1).
	// Aquí asumimos que solo se devuelve una línea (el primer proceso encontrado).
	fields := strings.Fields(string(output))
	if len(fields) < 2 {
		return 0, fmt.Errorf("no process found (or ps output unexpected) for '%s'", processName)
	}
	pidStr := fields[1]
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("error parsing ps output (%s): %v", pidStr, err)
	}
	return pid, nil
}
