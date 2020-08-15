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
	testFs = New(bpFs, "/", true).(*Afero) // debug true so failing tests leave more information

	result := m.Run()

	err = fs.RemoveAll(name)
	if err != nil {
		log.Println("Error removing test artifacts: ", err)
		os.Exit(1)
	}

	os.Exit(result)
}

