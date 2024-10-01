package filesystem

import (
	"io/ioutil"
	"os"
)

// IFileSystem interface to abstract file system operations
type IFileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	ReadFile(filename string) ([]byte, error)
	Remove(name string) error
	IsNotExist(err error) bool
}

type OSFileSystem struct{}

// Implement the methods on OSFileSystem to conform to FileSystem interface.
func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OSFileSystem) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (OSFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func (OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}
