package detector

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// SystemInfo contiene la información del sistema detectada
type SystemInfo struct {
	OS            string
	Distro        string
	PerfInstalled bool
	PerfVersion   string
}

// DetectSystem detecta información del sistema operativo y distribución
func DetectSystem() (*SystemInfo, error) {
	info := &SystemInfo{}

	// Detectar OS
	info.OS = "linux" // Por ahora asumimos Linux, podríamos expandir después

	// Detectar distribución
	if _, err := os.Stat("/etc/os-release"); err == nil {
		cmd := exec.Command("cat", "/etc/os-release")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("error leyendo /etc/os-release: %v", err)
		}

		// Parsear ID de distribución
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ID=") {
				info.Distro = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
				break
			}
		}
	}

	// Verificar si perf está instalado para el kernel actual
	kernelCmd := exec.Command("uname", "-r")
	kernelOut, err := kernelCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not determine kernel version: %v", err)
	}
	kernelVersion := strings.TrimSpace(string(kernelOut))
	perfPath := "/usr/lib/linux-tools-" + kernelVersion + "/perf"
	if _, err := os.Stat(perfPath); err == nil {
		info.PerfInstalled = true
		info.PerfVersion = perfPath
	} else if _, err := exec.LookPath("perf"); err == nil {
		info.PerfInstalled = true
		cmd := exec.Command("perf", "--version")
		output, err := cmd.Output()
		if err == nil {
			info.PerfVersion = strings.TrimSpace(string(output))
		}
	} else {
		info.PerfInstalled = false
		return nil, fmt.Errorf("perf is not installed for your kernel (%s). Please run: sudo apt-get install linux-tools-%s linux-cloud-tools-%s", kernelVersion, kernelVersion, kernelVersion)
	}

	return info, nil
}

// CheckPermissions verifica los permisos necesarios para perf
func CheckPermissions() error {
	// Verificar perf_event_paranoid
	contents, err := ioutil.ReadFile("/proc/sys/kernel/perf_event_paranoid")
	if err != nil {
		return fmt.Errorf("could not read /proc/sys/kernel/perf_event_paranoid: %v", err)
	}
	value := strings.TrimSpace(string(contents))
	if value != "-1" && value != "0" && value != "1" {
		return fmt.Errorf("Your system restricts performance monitoring (perf_event_paranoid=%s).\nTo allow perf, run: sudo sysctl -w kernel.perf_event_paranoid=1\nFor more info: https://www.kernel.org/doc/html/latest/admin-guide/perf-security.html", value)
	}
	return nil
}

// InstallPerf instala perf si no está presente
func InstallPerf(distro string) error {
	var cmd *exec.Cmd

	switch distro {
	case "ubuntu", "debian":
		cmd = exec.Command("sudo", "apt-get", "update")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error actualizando repositorios: %v", err)
		}
		cmd = exec.Command("sudo", "apt-get", "install", "-y", "linux-tools-common", "linux-tools-generic")
	case "fedora", "rhel", "centos":
		cmd = exec.Command("sudo", "dnf", "install", "-y", "perf")
	default:
		return fmt.Errorf("distribución no soportada: %s", distro)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error instalando perf: %v", err)
	}

	return nil
}
