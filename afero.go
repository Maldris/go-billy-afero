// Package afero provides a billy filesystem that wraps the afero api.
package afero // import "github.com/Maldris/go-billy-afero"

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-billy"
	"github.com/spf13/afero"
)

const (
	defaultDirectoryMode = 0755
	defaultCreateMode    = 0666
)

// Afero is a wrapper of the Afero API.
type Afero struct {
	fs    afero.Fs
	root  string
	Debug bool
}

// New returns a new OS filesystem.
func New(fs afero.Fs, root string, debug bool) billy.Filesystem {
	// TODO: rewrite this
	return &Afero{fs: fs, root: root, Debug: debug}
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (fs *Afero) Create(filename string) (billy.File, error) {
	if fs.Debug {
		log.Println("Create ", filename)
	}
	return fs.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultCreateMode)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (fs *Afero) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	if fs.Debug {
		log.Println("OpenFile ", filename)
	}
	if flag&os.O_CREATE != 0 {
		if err := fs.createDir(filename); err != nil {
			return nil, err
		}
	}

	f, err := fs.fs.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}
	name := filepath.ToSlash(f.Name())
	if strings.HasPrefix(name, fs.root) {
		name = strings.TrimPrefix(name, fs.root)
	}
	return &file{File: f, name: name}, err
}

func (fs *Afero) createDir(fullpath string) error {
	if fs.Debug {
		log.Println("createDir ", fullpath)
	}
	dir := path.Dir(fullpath)
	if dir != "." {
		if err := fs.MkdirAll(dir, defaultDirectoryMode); err != nil {
			return err
		}
	}

	return nil
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (fs *Afero) ReadDir(path string) ([]os.FileInfo, error) {
	if fs.Debug {
		log.Println("ReadDir ", path)
	}
	l, err := afero.ReadDir(fs.fs, path)
	if err != nil {
		return nil, err
	}

	var s = make([]os.FileInfo, len(l))
	for i, f := range l {
		s[i] = f
	}

	return s, nil
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (fs *Afero) Rename(from, to string) error {
	if fs.Debug {
		log.Println("Rename \n", from, "\n", to)
	}
	if err := fs.createDir(to); err != nil {
		return err
	}

	return fs.fs.Rename(from, to)
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func (fs *Afero) MkdirAll(path string, perm os.FileMode) error {
	if fs.Debug {
		log.Println("MkdirAll ", path)
	}
	return fs.fs.MkdirAll(path, defaultDirectoryMode)
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (fs *Afero) Open(filename string) (billy.File, error) {
	if fs.Debug {
		log.Println("Open ", filename)
	}
	return fs.OpenFile(filename, os.O_RDONLY, 0)
}

// Stat returns a FileInfo describing the named file.
func (fs *Afero) Stat(filename string) (os.FileInfo, error) {
	if fs.Debug {
		log.Println("Stat ", filename)
	}
	return fs.fs.Stat(filename)
}

// Remove removes the named file or directory.
func (fs *Afero) Remove(filename string) error {
	if fs.Debug {
		log.Println("Remove ", filename)
	}
	return fs.fs.Remove(filename)
}

// TempFile creates a new temporary file in the directory dir with a name
// beginning with prefix, opens the file for reading and writing, and
// returns the resulting *os.File. If dir is the empty string, TempFile
// uses the default directory for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously will not choose the
// same file. The caller can use f.Name() to find the pathname of the file.
// It is the caller's responsibility to remove the file when no longer
// needed.
func (fs *Afero) TempFile(dir, prefix string) (billy.File, error) {
	if fs.Debug {
		log.Println("TempFile ", dir, "\n", prefix)
	}
	if err := fs.createDir(dir + "/"); err != nil {
		return nil, err
	}

	f, err := afero.TempFile(fs.fs, dir, prefix)
	if err != nil {
		return nil, err
	}
	name := filepath.ToSlash(f.Name())
	if strings.HasPrefix(name, fs.root) {
		name = strings.TrimPrefix(name, fs.root)
	}
	if fs.Debug {
		log.Println("Tempfile created: ", name)
	}
	return &file{File: f, name: name}, nil
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (fs *Afero) Join(elem ...string) string {
	if fs.Debug {
		log.Println("Join ", elem)
	}
	return path.Join(elem...)
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (fs *Afero) RemoveAll(filePath string) error {
	if fs.Debug {
		log.Println("RemoveAll ", filePath)
	}
	return fs.fs.RemoveAll(path.Clean(filePath))
}

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link.
func (fs *Afero) Lstat(filename string) (os.FileInfo, error) {
	if fs.Debug {
		log.Println("Lstat ", filename)
	}
	if lstater, ok := fs.fs.(afero.Lstater); ok {
		fileInfo, _, err := lstater.LstatIfPossible(filename)
		return fileInfo, err
	}
	return fs.Stat(path.Clean(filename))
}

// Symlink creates a symbolic-link from link to target. target may be an
// absolute or relative path, and need not refer to an existing node.
// Parent directories of link are created as necessary.
func (fs *Afero) Symlink(target, link string) error {
	if fs.Debug {
		log.Println("Symlink ", target, "\n", link)
	}
	if err := fs.createDir(link); err != nil {
		return err
	}

	if linker, ok := fs.fs.(afero.Linker); ok {
		return linker.SymlinkIfPossible(target, link)
	}

	return &os.LinkError{Op: "symlink", Old: target, New: link, Err: afero.ErrNoSymlink}
}

// Readlink returns the target path of link.
func (fs *Afero) Readlink(link string) (string, error) {
	if fs.Debug {
		log.Println("Readlink ", link)
	}
	if reader, ok := fs.fs.(afero.LinkReader); ok {
		return reader.ReadlinkIfPossible(link)
	}

	return "", &os.PathError{Op: "readlink", Path: link, Err: afero.ErrNoReadlink}
}

// Chroot returns a new filesystem from the same type where the new root is
// the given path. Files outside of the designated directory tree cannot be
// accessed.
func (fs *Afero) Chroot(fPath string) (billy.Filesystem, error) {
	if fs.Debug {
		log.Println("Chroot ", fPath)
	}
	return &Afero{fs: afero.NewBasePathFs(fs.fs, fPath), root: path.Join(fs.root, fPath), Debug: fs.Debug}, nil
}

// Root returns the root path of the filesystem.
func (fs *Afero) Root() string {
	if fs.Debug {
		log.Println("Root")
	}
	return fs.root
}

// Capabilities implements the Capable interface.
func (fs *Afero) Capabilities() billy.Capability {
	if fs.Debug {
		log.Println("Capabilities")
	}
	return billy.DefaultCapabilities
}

// file is a wrapper for an os.File which adds support for file locking.
type file struct {
	afero.File
	name string
	m    sync.Mutex
}

// Lock requests that a file is lock
func (f *file) Lock() error {
	f.m.Lock()
	return nil
}

// Unlock
func (f *file) Unlock() error {
	f.m.Unlock()
	return nil
}

func (f *file) Name() string {
	return f.name
}
