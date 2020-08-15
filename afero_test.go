package afero

import (
	"log"
	"os"
	"testing"

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
