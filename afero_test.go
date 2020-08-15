package afero

import (
	"log"
	"os"
	"testing"

	"github.com/spf13/afero"
)

var testFs *Afero

var (
	rootFileCont   = "I'm in the root path"
	dirFileCont1   = "I'm in a directory"
	dirFileCont2   = "" // file exists but is empty
	dirFileCont3   = "also in a directory"
	nestedFileCont = "I'm in a deeply nested path"
)

func TestMain(m *testing.M) {
	fs := afero.NewOsFs()
	name, err := afero.TempDir(fs, "/", "tests")
	if err != nil {
		log.Println("Error creating temp directory for testing: ", err)
		os.Exit(1)
	}
	bpFs := afero.NewBasePathFs(fs, name)
	err = createTestFileset(bpFs)
	if err != nil {
		log.Println("Error creating test fileset: ", err)
		os.Exit(1)
	}
	testFs = New(bpFs, "/", true).(*Afero) // debug true so failing tests leave more information

	result := m.Run()

	err = fs.RemoveAll(name)
	if err != nil {
		log.Println("Error removing test artifacts: ", err)
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

	err = afero.WriteFile(fs, "dir/deleteMe", []byte(dirFileCont1), defaultCreateMode)
	if err != nil {
		return errors.Wrap(err, "Error creating test file 'dir/deleteMe'")
	}

	return nil
}

