package host

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

const (
	defaultCPUInfoFile = "cpuinfo"
	procFolder         = "proc"
	cpuInfo1           = "hack/test/data/proc/cpuinfo-1"
	testDataDir        = "hack/test/data"
	testRunDir         = "hack/test/run"
)

func TestGetHardware(t *testing.T) {
	err := newTestRun()
	if err != nil {
		t.Logf("failed to prepare test case. Error was: %s", err)
		t.Fail()
	}
	generatedProcPath, err := createMockProc()
	if err != nil {
		t.Logf("failed to create mock proc dir. Error was: %s", err)
		t.Fail()
	}
	lr := NewLinuxReader(LinuxReaderConfig{
		ProcDirPath: *generatedProcPath,
	})
	hw, err := lr.GetHardware()
	if err != nil {
		t.Logf("failed to make GetHardware call. Error was: %s", err)
		t.Fail()
	}
	if hw.CPU.CPUCount != 8 {
		t.Logf("failed valid CPU count check. expected: %d, actual: %d.", 8, hw.CPU.CPUCount)
		t.Fail()
	}
}

// createMockProc creates a mock proc directory with contents and returns the
// location the proc direction was created at.
func createMockProc() (*string, error) {
	dir, err := os.MkdirTemp(testRunDir, "*")
	if err != nil {
		return nil, err
	}
	generatedProcPath := filepath.Join(dir, procFolder)
	err = os.Mkdir(generatedProcPath, 0777)
	if err != nil {
		return nil, err
	}
	err = addCPUInfoFile(dir, cpuInfo1)
	if err != nil {
		return nil, err
	}
	return &generatedProcPath, nil
}

func addCPUInfoFile(testDir, cpuInfoFile string) error {
	cpuInfoDataFile, err := os.Open(cpuInfoFile)
	if err != nil {
		return err
	}
	defer cpuInfoDataFile.Close()
	log.Println(filepath.Join(testDir, procFolder, defaultCPUInfoFile))
	cpuInfoMockFile, err := os.Create(filepath.Join(testDir, procFolder, defaultCPUInfoFile))
	if err != nil {
		return err
	}
	defer cpuInfoMockFile.Close()
	_, err = io.Copy(cpuInfoMockFile, cpuInfoDataFile)
	if err != nil {
		return err
	}

	return nil
}

// newTestRun ensures the testRunDir is created. Before attempting creation, it
// will also run [cleanTestRun] to ensure any existing content is removed.
func newTestRun() error {
	cleanTestRun()
	err := os.MkdirAll(testRunDir, 0777)
	if err != nil {
		return err
	}
	return nil
}

// cleanTestData remove any contents inside of hack/test/run.
// This can be called before new tests are run.
func cleanTestRun() error {
	err := os.RemoveAll(testRunDir)
	if err != nil {
		return err
	}
	return nil
}
