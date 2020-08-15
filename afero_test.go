package afero

import (
	"io/ioutil"
	"log"
	"os"
	"testing"


	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var (
	testFs  *Afero
	tempDir string
)

const (
	rootFileCont   = "I'm in the root path"
	dirFileCont1   = "I'm in a directory"
	dirFileCont2   = "" // file exists but is empty
	dirFileCont3   = "also in a directory"
	nestedFileCont = "I'm in a deeply nested path"
)

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Println("Errorgetting working directory: ", err)
		os.Exit(1)
	}
	fs := afero.NewOsFs()
	name, err := afero.TempDir(fs, pwd, "tests.")
	if err != nil {
		log.Println("Error creating temp directory for testing: ", err)
		os.Exit(1)
	}
	tempDir = name

	bpFs := afero.NewBasePathFs(fs, name)
	err = createTestFileset(bpFs)
	if err != nil {
		log.Println("Error creating test fileset: ", err)
		err = fs.RemoveAll(name)
		if err != nil {
			log.Println("Error deleting test directory: ", err)
		}
		os.Exit(1)
	}
	testFs = New(bpFs, name, true).(*Afero) // debug true so failing tests leave more information

	result := m.Run()

	err = fs.RemoveAll(name)
	if err != nil {
		log.Println("Error deleting test directory: ", err)
		os.Exit(1)
	}

	os.Exit(result)
}

func createTestFileset(fs afero.Fs) error {
	err := fs.Mkdir("dir", defaultDirectoryMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test directory 'dir'")
	}

	err = fs.MkdirAll("nested/test/dir", defaultDirectoryMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test path 'nested/test/dir'")
	}

	err = fs.MkdirAll("dir/nested/test/folder", defaultDirectoryMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test path 'dir/nested/test/folder'")
	}

	err = afero.WriteFile(fs, "root.file", []byte(rootFileCont), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'root.file'")
	}

	err = afero.WriteFile(fs, "dir/file1", []byte(dirFileCont1), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/file1'")
	}

	err = afero.WriteFile(fs, "dir/file.2", []byte(dirFileCont2), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/file.2'")
	}

	err = afero.WriteFile(fs, "dir/3file", []byte(dirFileCont3), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/3file'")
	}

	err = afero.WriteFile(fs, "nested/test/dir/file", []byte(nestedFileCont), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'nested/test/dir/file'")
	}

	err = afero.WriteFile(fs, "dir/nested/deleteMe", []byte(dirFileCont1), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/nested/deleteMe'")
	}

	err = afero.WriteFile(fs, "dir/nested/renameMe", []byte(dirFileCont1), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/nested/renameMe'")
	}

	err = afero.WriteFile(fs, "dir/nested/test/folder/file1", []byte(dirFileCont1), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/nested/test/folder/file1'")
	}

	err = afero.WriteFile(fs, "dir/nested/test/folder/file2", []byte(dirFileCont2), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/nested/test/folder/file2'")
	}

	err = afero.WriteFile(fs, "dir/nested/test/folder/file3", []byte(dirFileCont3), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/nested/test/folder/file3'")
	}

	linker := fs.(afero.Symlinker)

	err = linker.SymlinkIfPossible("dir/file1", "dir/nested/test/symlink")
	if err != nil {
		return errors.Wrap(err, "Error creating symlink from 'dir/file1' to 'dir/nested/test/symlink'")
	}

	return nil
}

// ================
// Filesystem Tests
// ================

func TestCreate(t *testing.T) {
	f, err := testFs.Create("rootFile")
	if err != nil {
		t.Error("Error creating test file: ", err)
		return
	}
	defer f.Close()

	st, err := testFs.fs.Stat("rootFile")
	if err != nil {
		t.Error("Unable to stat created file: ", err)
		return
	}
	if st == nil {
		t.Error("Test file does not exist")
		return
	}

	n, err := f.Write([]byte(dirFileCont1))
	if err != nil {
		t.Error("Error writing to created file: ", err)
		return
	}
	if n != len(dirFileCont1) {
		t.Error("write length does not match content length: ", n, " of ", len(dirFileCont1))
		return
	}

	data, err := afero.ReadFile(testFs.fs, "rootFile")
	if err != nil {
		t.Error("Error reading created file content: ", err)
		return
	}

	if string(data) != dirFileCont1 {
		t.Error("File content does not match written value: ", err)
		return
	}
}

func TestCreate2(t *testing.T) {
	f, err := testFs.Create("dir/file.2")
	if err != nil {
		t.Error("Error creating test file: ", err)
		return
	}
	defer f.Close()

	n, err := f.Write([]byte(dirFileCont1))
	if err != nil {
		t.Error("Error writing to truncated file: ", err)
		return
	}
	if n != len(dirFileCont1) {
		t.Error("write length does not match content length: ", n, " of ", len(dirFileCont1))
		return
	}

	data, err := afero.ReadFile(testFs.fs, "dir/file.2")
	if err != nil {
		t.Error("Error reading truncated file content: ", err)
		return
	}

	if string(data) != dirFileCont1 {
		t.Error("File content does not match written value: ", err)
		return
	}
}

func TestOpenFile(t *testing.T) {
	f, err := testFs.OpenFile("root.file", os.O_RDONLY, 0)
	if err != nil {
		t.Error("Error opening file: ", err)
		return
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error("Error reading file content: ", err)
		return
	}

	if string(content) != rootFileCont {
		t.Error("File content does not match expected test value")
		return
	}
}

func TestReadDir(t *testing.T) {
	sts, err := testFs.ReadDir("dir")
	if err != nil {
		t.Error("Error reading directory: ", err)
		return
	}

	if len(sts) != 4 {
		t.Error("Not the expected number of files found")
		return
	}

	type statResult struct {
		found bool
		isDir bool
	}

	found := map[string]statResult{
		"file1":  statResult{found: false, isDir: false},
		"file.2": statResult{found: false, isDir: false},
		"3file":  statResult{found: false, isDir: false},
		"nested": statResult{found: false, isDir: true},
	}

	for _, st := range sts {
		expect := found[st.Name()]
		if st.IsDir() != expect.isDir {
			t.Error(st.Name() + " does not match the expected file/directory status")
		}
		expect.found = true
		found[st.Name()] = expect
	}

	for name := range found {
		if !found[name].found {
			t.Error("Expected file " + name + " not found")
		}
	}
}

func TestRename(t *testing.T) {
	err := testFs.Rename("dir/nested/renameMe", "dir/nested/renamed")
	if err != nil {
		t.Error("Error renaming file: ", err)
		return
	}

	data, err := afero.ReadFile(testFs.fs, "dir/nested/renamed")
	if err != nil {
		t.Error("Error reading renamed file", err)
		return
	}

	if string(data) != dirFileCont1 {
		t.Error("Renamed file content is not that of original file")
	}
}

func TestMkdirAll(t *testing.T) {
	err := testFs.MkdirAll("make/this/directory", defaultDirectoryMode)
	if err != nil {
		t.Error("Error making all directories to path: ", err)
		return
	}

	st, err := testFs.fs.Stat("make/this/directory")
	if err != nil {
		t.Error("Error stating created directory: ", err)
		return
	}

	if st == nil {
		t.Error("Stat of created directory is nil")
		return
	}

	if !st.IsDir() {
		t.Error("Created directory is not a directory")
		return
	}
}

func TestOpen(t *testing.T) {
	f, err := testFs.Open("dir/3file")
	if err != nil {
		t.Error("Error opening file: ", err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error("Error reading file content: ", err)
		return
	}

	if string(data) != dirFileCont3 {
		t.Error("File content does not match expected value")
	}
}

func TestStat(t *testing.T) {
	st, err := testFs.Stat("nested/test/dir/file")
	if err != nil {
		t.Error("Unable to stat test file: ", err)
		return
	}

	if st == nil {
		t.Error("Stat of test file should not be nil")
		return
	}

	if st.IsDir() {
		t.Error("Stat reports file as directory")
		return
	}

	if st.Size() != int64(len(nestedFileCont)) {
		t.Error("File size does not match assigned content")
	}
}

func TestStat2(t *testing.T) {
	st, err := testFs.Stat("nested/test/dir")
	if err != nil {
		t.Error("Unable to stat test directory: ", err)
		return
	}

	if st == nil {
		t.Error("Stat of test directory should not be nil")
		return
	}

	if !st.IsDir() {
		t.Error("Stat reports directory as file")
		return
	}
}

func TestRemove(t *testing.T) {
	err := testFs.Remove("dir/nested/deleteMe")
	if err != nil {
		t.Error("Error while deleting test file: ", err)
		return
	}

	_, err = testFs.fs.Stat("dir/nested/deleteMe")
	if err == nil {
		t.Error("Stat of deleted file succeeded")
		return
	}
	// TODO: how best to check its the correct error?
}

func TestRemove2(t *testing.T) {
	err := testFs.Remove("dir/nested/non-existant")
	if err == nil {
		t.Error("Deleting a non-existant file succeeded")
		return
	}
	// TODO: how best to check its the correct error?
}

func TestTempFile(t *testing.T) {
	f, err := testFs.TempFile("dir/nested/test", "temp")
	if err != nil {
		t.Error("Error creating temp file: ", err)
		return
	}
	defer f.Close()

	_, err = testFs.fs.Stat(f.Name())
	if err != nil {
		t.Error("Error stating new temp file: ", err)
	}

	// don't worry about deleting the temp file, will be handled in test process cleanup
}

func TestJoin(t *testing.T) {
	set := map[string][]string{
		"test/join":                 {"test", "join"},
		"test/longer/join":          {"test", "longer/join"},
		"test/join/fragment":        {"test/join", "fragment"},
		"test/longer/join/fragment": {"test/longer", "join/fragment"},
		"join/test":                 {"join", "", "", "", "test"},
		"clean/join":                {"clean///", "join"},
		"/absolute/join":            {"///absolute////", "", "", "////join////"},
	}

	for result, inp := range set {
		res := testFs.Join(inp...)
		if result != res {
			t.Error("Join did not produce the expected result of '"+result+"' for the input: ", inp, " instead got: ", res)
		}
	}
}

