package host

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

const (
	defaultCPUInfoFile   = "cpuinfo"
	defaultMachineIDFile = "machine-id"
	procFolder           = "proc"
	etcFolder            = "etc"
	cpuInfo1             = "hack/test/data/proc/cpuinfo-1"
	machineID1           = "hack/test/data/etc/machine-id-1"
	testDataDir          = "hack/test/data"
	testRunDir           = "hack/test/run"
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

func TestGetHostID(t *testing.T) {
	err := newTestRun()
	if err != nil {
		t.Logf("failed to prepare test case. Error was: %s", err)
		t.Fail()
	}
	mIDPath, err := createMockMachineID()
	if err != nil {
		t.Logf("failed setting up mock machineID file. Error was: %s", err)
		t.FailNow()
	}
	lr := NewLinuxReader(LinuxReaderConfig{
		MachineIDPath: *mIDPath,
	})
	id, err := lr.GetHostID()
	if err != nil {
		t.Logf("failed resolving machine id. Error was: %s", err)
		t.FailNow()
	}
	expectedMachineID := "abc123xyz"
	if id != expectedMachineID {
		t.Logf("failed with unexpected machine id. Expected: %s, actual: %s", expectedMachineID, id)
		t.Fail()
	}
}

func createMockMachineID() (*string, error) {
	dir, err := os.MkdirTemp(testRunDir, "*")
	if err != nil {
		return nil, err
	}
	generatedMachineIDPath := filepath.Join(dir, etcFolder)
	err = os.Mkdir(generatedMachineIDPath, 0777)
	if err != nil {
		return nil, err
	}
	addMachineIDFile(generatedMachineIDPath, machineID1)

	path := filepath.Join(generatedMachineIDPath, defaultMachineIDFile)
	return &path, nil
}

func addMachineIDFile(testDir, machineIDFile string) error {
	machineIDDataFile, err := os.Open(machineIDFile)
	if err != nil {
		return err
	}
	defer machineIDDataFile.Close()

	log.Println(filepath.Join(testDir, defaultMachineIDFile))
	machineIDMockFile, err := os.Create(filepath.Join(testDir, defaultMachineIDFile))
	if err != nil {
		return err
	}
	defer machineIDMockFile.Close()
	_, err = io.Copy(machineIDMockFile, machineIDDataFile)
	if err != nil {
		return err
	}

	return nil

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
