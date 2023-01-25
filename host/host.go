// The host package is responsible for gathering details about a given host.
package host

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	DefaultMachineIDPath = "/etc/machine-id"
	DefaultProcRoot      = "/proc"
	OSReleaseFilePath    = "/etc/os-release"
	OSKernelFilePath     = "sys/kernel/osrelease"
	CPUInfoFilePath      = "cpuinfo"
	UnknownKey           = "UNKNOWN"
)

// OS represents details about the operating system.
type OS struct {
	Name    string
	Version string
}

// Kernel represents the operating-system's kernel's details.
type Kernel struct {
	Type    string
	Version string
}

// Hardware represents the hardware on the machine.
type Hardware struct {
	CPU          CPUInfo
	Architecture string
}

// CPUInfo represents details about the central processing unit.
type CPUInfo struct {
	CPUCount int
}

// HostReader defines the actions available for retrieving information about a host.
type HostReader interface {
	// GetOS retrieves operating-system details
	GetOS() (*OS, error)
	// GetKernel retrieves kernel details.
	GetKernel() (*Kernel, error)
	// GetHardware retrieves hardware-level details. Or, in the case of a virtual machine, what is
	// exposed to the guest.
	GetHardware() (*Hardware, error)
	// GetHostID retrieves a unique identifier that represents the host (physical/virtual machine).
	GetHostID() (string, error)
}

// LinuxReader is the Linux-specific implementation of [HostReader].
type LinuxReader struct {
	procDir       string
	machineIDPath string
}

type LinuxReaderConfig struct {
	ProcDirPath   string
	MachineIDPath string
}

func NewLinuxReader(conf LinuxReaderConfig) LinuxReader {
	if conf.ProcDirPath == "" {
		conf.ProcDirPath = DefaultProcRoot
	}
	if conf.MachineIDPath == "" {
		conf.MachineIDPath = DefaultMachineIDPath
	}
	return LinuxReader{
		procDir:       conf.ProcDirPath,
		machineIDPath: conf.MachineIDPath,
	}
}

// GetOS looks up details about the operating system within /etc/os-release.
// We rely on details found inside os-release that comply with metadata found in the [freedesktop
// specification].
//
// [freedesktop specification]: https://www.freedesktop.org/software/systemd/man/os-release.html
func (h *LinuxReader) GetOS() (*OS, error) {
	releaseFileData, err := os.ReadFile(OSReleaseFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed locating OS details at %s. Error was: %s",
			OSReleaseFilePath, err)
	}

	OSReleaseData := parseOSRelease(releaseFileData)
	return &OS{
		Name:    OSReleaseData["ID"],
		Version: OSReleaseData["VERSION"],
	}, nil
}

// GetKernel retrieves details about the kernel of the operating system.
func (h *LinuxReader) GetKernel() (*Kernel, error) {
	kernelFilePath := filepath.Join(h.procDir, OSKernelFilePath)
	kernelFileData, err := os.ReadFile(kernelFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed getting kernel version from %s. Error was: %s", OSKernelFilePath, err)
	}
	fmt.Printf("kernel version: %s\n\n", kernelFileData)
	return &Kernel{
		Type:    "Linux",
		Version: string(kernelFileData),
	}, nil
}

func (h *LinuxReader) GetHardware() (*Hardware, error) {
	arch := getArch()
	CPUInfo := h.getCPUInfo()

	return &Hardware{
		CPU:          CPUInfo,
		Architecture: arch,
	}, nil
}

// GetHostID provides a unique identifier representing the host. Today, it relies on [machine-id],
// which is created by Linux during installation. This functionality can be expanded over time to
// add methods for detecting the ID, when /etc/machine-id isn't possible or inadequate. If a ID is
// unable to be resolved, an error is returned.
func (h *LinuxReader) GetHostID() (string, error) {
	midBytes, err := os.ReadFile(h.machineIDPath)
	if err != nil {
		return "", fmt.Errorf("failed resolving machine ID. Error was: %s", err)
	}
	// machineID not present in file; this is an error
	if len(midBytes) < 1 {
		return "", fmt.Errorf("failed resolving machine ID. Error was: machine-id file present but empty.")
	}
	return string(midBytes), nil
}

// getCPUInfo retrieves details about the system's CPU based on /proc/cpuinfo.
// TOOD(joshrosso): Right now we just get CPU count, this can be expanded for more details, such as
// clock speed. If there's an error reading necessary files, an empty CPU Info is returned.
func (h *LinuxReader) getCPUInfo() CPUInfo {
	processorCount := 0
	cpuInfoPath := filepath.Join(h.procDir, CPUInfoFilePath)
	f, err := os.Open(cpuInfoPath)
	if err != nil {
		log.Printf("failed retrieving processor type from %s. Error was: %s", CPUInfoFilePath, err)
		return CPUInfo{}
	}
	scanner := bufio.NewScanner(bufio.NewReader(f))
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		kv[0] = strings.TrimSpace(kv[0])
		kv[1] = strings.TrimSpace(kv[1])
		if kv[0] == "processor" {
			processorCount++
		}
	}
	return CPUInfo{
		CPUCount: processorCount,
	}
}

// getArch call the equivalent of uname -m to get the architecture (e.g. x86 or aarch64)
func getArch() string {
	var utsname unix.Utsname
	err := unix.Uname(&utsname)
	if err != nil {
		return UnknownKey
	}
	return string(utsname.Machine[:])
}

// sanitizeOSVersion removes a double quote character from the beginning and end of a string if
// present.
func sanitizeOSVersion(version string) string {
	return strings.Trim(version, "\"")
}

// parseOSRelease takes the contents of an /etc/os-release file and returns a map containing each
// key/value pair. The key/value pair is determined by parsing the syntax of $KEY=$VALUE within the
// file.
func parseOSRelease(releaseFileContents []byte) map[string]string {
	scanner := bufio.NewScanner(bytes.NewReader(releaseFileContents))
	osReleaseMap := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			osReleaseMap[kv[0]] = kv[1]
		}
	}
	return osReleaseMap
}
