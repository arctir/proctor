package plib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

const (
	DefaultFilePerms = 0777
	StatDirName      = "stat"
	HackDir          = "hack"
	TestingDir       = "test"
	TestDataDir      = "data-dir"
	TestProcDir      = "proc"
	TestCacheDir     = "cache"
	TestCacheFile    = "proc.cache"
	StatData1002     = `1002 (Thunar) S 898 898 898 0 -1 4194304 9075 31619 19 0 242 54 42 7 20 0 3 0 4316 499617792 14545 18446744073709551615 94657007656960 94657008059597 140727172487872 0 0 0 0 4096 0 0 0 0 17 10 0 0 0 0 0 94657008206176 94657008240992 94657028120576 140727172496280 140727172496349 140727172496349 140727172497384 0`
	StatData68657    = `68657 (chromium) S 68654 68650 68650 0 -1 4194560 1462096 116023 16 0 13834 4693 47 34 20 0 23 0 7679775 35172757504 80617 18446744073709551615 94708279918592 94708471088624 140731884479632 0 0 0 0 4096 1098990847 0 0 0 17 0 0 0 0 0 0 94708479643648 94708480167272 94708482572288 140731884485166 140731884485225 140731884485225 140731884486621 0`
)

func TestLoadProcesses(t *testing.T) {
	procFp := getTestProcDir()
	cacheFp := getTestCacheDir()
	err := createDirsAndSampleData()
	if err != nil {
		t.Fatalf("failed setting up sample data for test: %s", err)
	}
	defer cleanTestData()

	// verify an error is returned when the location of procfs doesn't exist
	badProcFsPath := filepath.Join("hack", "fake", "path")
	config := LinuxInspectorConfig{
		ProcfsFilePath: badProcFsPath,
		CacheFilePath:  cacheFp,
	}
	li := NewLinuxInspector(config)
	err = li.LoadProcesses()
	if err == nil {
		t.Logf("error was expected since procfs (%s) is not a real location. However no error was returned.", badProcFsPath)
		t.Fail()
	}

	// verify no error is returned when the procfs location is valid
	config2 := LinuxInspectorConfig{
		ProcfsFilePath: procFp,
		CacheFilePath:  cacheFp,
	}
	li2 := NewLinuxInspector(config2)
	err = li2.LoadProcesses()
	if err != nil {
		t.Logf("error unexpectadly returned when the procfs location (%s) and data was valid. error from call: %s", procFp, err)
		t.Fail()
	}

	// verify data is as expected in in-memory cache
	pidVal := 1002
	if li2.ps[pidVal].ID != pidVal {
		t.Logf("error process had invalid data. For process with key %d, expected: %d actual: %d", pidVal, pidVal, li2.ps[pidVal].ID)
		t.Fail()
	}

	pidVal2 := 68657
	if li2.ps[pidVal2].ID != pidVal2 {
		t.Logf("error process had invalid data. For process with key %d, expected: %d actual: %d", pidVal2, pidVal2, li2.ps[pidVal2].ID)
		t.Fail()
	}

	// verify cache is filled
	cacheFilePath := filepath.Join(cacheFp, CacheFileName)
	_, err = os.Stat(cacheFilePath)
	if os.IsNotExist(err) {
		t.Logf("expected cache file at %s, but did not find one.", cacheFilePath)
		t.Fail()
	}
}

func TestClearProcessCache(t *testing.T) {
	procFp := getTestProcDir()
	cacheFp := getTestCacheDir()
	err := createDirsAndSampleData()
	if err != nil {
		t.Fatalf("failed setting up sample data for test: %s", err)
	}
	defer cleanTestData()

	// run load LoadProcessess
	config := LinuxInspectorConfig{
		ProcfsFilePath: procFp,
		CacheFilePath:  cacheFp,
	}
	li := NewLinuxInspector(config)
	err = li.LoadProcesses()
	if err != nil {
		t.Logf("error unexpectadly returned when the procfs location (%s) and data was valid. error from call: %s", procFp, err)
		t.Fail()
	}

	// verify cache file is present
	_, err = os.Stat(filepath.Join(cacheFp, CacheFileName))
	// if there is any form of error stating the file, then the're an issue with
	// the cache file or it does not exist. Thus, the test should fail.
	if err != nil {
		t.Fatalf("failed attempting to verify the existence of the cache file (via stat): %s", err)
	}

	// run clear cache
	err = li.ClearProcessCache()
	if err != nil {
		t.Fatalf("an errror was returned while attempting to clear the process cache: %s", err)
	}

	// verify cache is removed
	_, err = os.Stat(filepath.Join(cacheFp, CacheFileName))
	// if no error is returned, then the file exists, which means the clear did
	// not work. In this case, it's a test failure.
	if err == nil {
		t.Fatalf("was able to locate the cache file (%s) even though it was supposed to clear.", cacheFp)
	}
	// additionally, if an error did occur, but it's not becuase the file did not
	// exist, then that's also a failure.
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("an error occured trying to verify the cache file was cleared: %s", err)
	}
}

func TestGetProcesses(t *testing.T) {
	procFp := getTestProcDir()
	cacheFp := getTestCacheDir()
	err := createDirsAndSampleData()
	if err != nil {
		t.Fatalf("failed setting up sample data for test: %s", err)
	}
	defer cleanTestData()

	// run load LoadProcessess
	config := LinuxInspectorConfig{
		ProcfsFilePath: procFp,
		CacheFilePath:  cacheFp,
	}
	li := NewLinuxInspector(config)
	ps, err := li.GetProcesses()
	if err != nil {
		t.Fatalf("failed retrieving processes: %s", err)
	}

	if ps[1002].CommandName != "(Thunar)" {
		t.Logf("command name for process %d was: %s but we expected %s", 1002, ps[1002].CommandName, "(Thunar)")
		t.Fail()
	}

	if ps[1002].ID != 1002 {
		t.Logf("pid for process %d was: %d but we expected %d", 1002, ps[1002].ID, 1002)
		t.Fail()
	}

	if ps[68657].CommandName != "(chromium)" {
		t.Logf("command name for process %d was: %s but we expected %s", 68657, ps[68657].CommandName, "(chromium)")
		t.Fail()
	}

	if ps[68657].ID != 68657 {
		t.Logf("pid for process %d was: %d but we expected %d", 68657, ps[68657].ID, 68657)
		t.Fail()
	}

}

// TestGetProcessesFromMemory loads all processes into memory (in the
// LinuxInspector) struct and then clears out the test-data dir containing the
// mock procfs and cache file. This ensures that in the absence of these, when
// the in-memory representation is filled, it can still return processes to the
// caller.
func TestGetProcessesFromMemory(t *testing.T) {
	procFp := getTestProcDir()
	cacheFp := getTestCacheDir()
	err := createDirsAndSampleData()
	if err != nil {
		t.Fatalf("failed setting up sample data for test: %s", err)
	}
	// even though this function calls clearTestData later, we still make this
	// defer call to ensure clean-up always happens (even if that means it
	// happens again) when this function exits.
	defer cleanTestData()

	// run load LoadProcessess
	config := LinuxInspectorConfig{
		ProcfsFilePath: procFp,
		CacheFilePath:  cacheFp,
	}
	li := NewLinuxInspector(config)
	err = li.LoadProcesses()
	if err != nil {
		t.Logf("error unexpectadly returned when the procfs location (%s) and data was valid. error from call: %s", procFp, err)
		t.Fail()
	}

	// clean-up any test-data (e.g. procfs and cache
	cleanTestData()

	// verify test data is non-existent
	_, err = os.Stat(cacheFp)
	// if no error is returned, then the file exists, which means the clear did
	// not work. In this case, it's a test failure.
	if err == nil {
		t.Fatalf("was able to locate the cache dir (%s) even though it was supposed to clear.", cacheFp)
	}
	_, err = os.Stat(procFp)
	// if no error is returned, then the file exists, which means the clear did
	// not work. In this case, it's a test failure.
	if err == nil {
		t.Fatalf("was able to locate the mock procfs dir (%s) even though it was supposed to clear.", cacheFp)
	}

	// make GetProcesses call
	ps, err := li.GetProcesses()
	if err != nil {
		t.Fatalf("failed retrieving processes: %s", err)
	}

	// ensure it contains processes (size)
	if len(ps) != 2 {
		t.Fatalf("%d processes were returned, when we expected there to be %d.", len(ps), 2)
	}

	// check a value
	if ps[1002].CommandName != "(Thunar)" {
		t.Logf("command name for process %d was: %s but we expected %s", 1002, ps[1002].CommandName, "(Thunar)")
		t.Fail()
	}

}

// TestGetProcessesFromFileCache puts a Get call in a situation where the
// processes have been loaded from procfs and cached to to filesystem.
// Subsequently, the in-memory process (l.ps) is set to nil and procfs is
// removed. This will make the get call only able to succeed if it can read
// from the file-based cache and return the results to the caller.
func TestGetProcessesFromFileCache(t *testing.T) {
	procFp := getTestProcDir()
	cacheFp := getTestCacheDir()
	err := createDirsAndSampleData()
	if err != nil {
		t.Fatalf("failed setting up sample data for test: %s", err)
	}
	// even though this function calls clearTestData later, we still make this
	// defer call to ensure clean-up always happens (even if that means it
	// happens again) when this function exits.
	defer cleanTestData()

	// run load LoadProcessess
	config := LinuxInspectorConfig{
		ProcfsFilePath: procFp,
		CacheFilePath:  cacheFp,
	}
	li := NewLinuxInspector(config)
	err = li.LoadProcesses()
	if err != nil {
		t.Logf("error unexpectadly returned when the procfs location (%s) and data was valid. error from call: %s", procFp, err)
		t.Fail()
	}

	// clear the in-memory process representation
	li.ps = nil

	// remove procfs from the test-data so no re-lookups can occur
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
	}
	fp := filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir, TestProcDir)
	err = os.RemoveAll(fp)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
		}
	}

	// verify procfs is removed
	_, err = os.Stat(procFp)
	// if no error is returned, then the file exists, which means the clear did
	// not work. In this case, it's a test failure.
	if err == nil {
		t.Fatalf("expected the mock procfs to be removed but was able to stat it at %s, making this testing invalid", procFp)
	}

	// verify cache exists
	_, err = os.Stat(cacheFp)
	// if no error is returned, then the file exists, which means the clear did
	// not work. In this case, it's a test failure.
	if err != nil {
		t.Fatalf("expected the cache to exist at %s, but could not stat it, making this test invalid", cacheFp)
	}

	// make GetProcesses call
	ps, err := li.GetProcesses()
	if err != nil {
		t.Fatalf("failed retrieving processes: %s", err)
	}

	// ensure it contains processes (size)
	if len(ps) != 2 {
		t.Fatalf("%d processes were returned, when we expected there to be %d.", len(ps), 2)
	}

	// check a value
	if ps[68657].CommandName != "(chromium)" {
		t.Logf("command name for process %d was: %s but we expected %s", 68657, ps[68657].CommandName, "(chromium)")
		t.Fail()
	}

}

func createDirsAndSampleData() error {
	procFp, err := createMockProcDir()
	if err != nil {
		return err
	}
	sampleData := []struct {
		pid  string
		data string
	}{
		{"1002", StatData1002},
		{"68657", StatData68657},
	}

	// Load all sample data for procfs
	for _, p := range sampleData {
		// pfp is the pid's filepath in proc
		pfp := filepath.Join(procFp, p.pid)
		err := os.MkdirAll(pfp, DefaultFilePerms)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(pfp, StatDirName), []byte(p.data), DefaultFilePerms)
		if err != nil {
			return err
		}
	}
	return nil
}

func createMockProcDir() (string, error) {
	fp := getTestProcDir()
	if _, err := os.Stat(fp); err != nil {
		// if the dir was stat'd (it exists) then remove it.
		if err == nil {
			err = os.Remove(fp)
			// return error if unable to remove existing file
			if err != nil {
				return "", fmt.Errorf("failed cleaning existing testing data directory: %s", err)
			}
		}
	}

	err := os.MkdirAll(fp, DefaultFilePerms)
	if err != nil {
		return "", fmt.Errorf("failed creating testing data directory: %s", err)
	}

	return fp, nil
}

func createCacheDir() (string, error) {
	fp := getTestCacheDir()
	if _, err := os.Stat(fp); err != nil {
		// if the dir was stat'd (it exists) then remove it.
		if err == nil {
			err = os.Remove(fp)
			// return error if unable to remove existing file
			if err != nil {
				return "", fmt.Errorf("failed cleaning existing testing data directory: %s", err)
			}
		}
	}

	err := os.MkdirAll(fp, DefaultFilePerms)
	if err != nil {
		return "", fmt.Errorf("failed creating testing data directory: %s", err)
	}

	return fp, nil
}

func getTestProcDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir, TestProcDir)
}

func getTestCacheDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir, TestCacheDir)
}

func cleanTestData() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
	}
	fp := filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir)
	err = os.RemoveAll(fp)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
		}
	}
}
