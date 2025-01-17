package config

import "os"

type FileSystem interface {
	Stat(name string) (os.FileInfo, error)
	Getwd() (string, error)
	Mkdir(name string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	Open(name string) (*os.File, error)
}

type realFileSystem struct{}

func (realFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (realFileSystem) Getwd() (string, error) {
	return os.Getwd()
}
func (realFileSystem) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}
func (realFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (realFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

type Options struct {
	FileSystem FileSystem
}

func DefaultOptions() Options {
	return Options{
		FileSystem: realFileSystem{},
	}
}

// Option is a functional option for configuring InitializeProject
type Option func(*Options)

// WithFileSystem allows overriding the FileSystem dependency. Only used for testing
func WithFileSystem(fs FileSystem) Option {
	return func(o *Options) {
		o.FileSystem = fs
	}
}
