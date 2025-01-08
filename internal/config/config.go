package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type ConfigError string

func (e ConfigError) Error() string {
	return string(e)
}

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

type ConfigFormat string

const (
	ConfigYaml ConfigFormat = "yaml"
	ConfigJson ConfigFormat = "json"
)

type Project struct {
	Name string `json:"name" yaml:"name"`
	ID   string `json:"id" yaml:"id"`
}

type Manifest struct {
	rootDir  string
	format   ConfigFormat
	TenantID string  `json:"tenantId" yaml:"tenantId"`
	Project  Project `json:"project" yaml:"project"`
}

func (c *Manifest) FunctionDir() string {
	return path.Join(c.rootDir, FunctionsDir)
}
func (c *Manifest) DataDir() string {
	return path.Join(c.rootDir, DataDir)
}

func (c *Manifest) ConfigPath() string {
	return path.Join(c.rootDir, DeploymentManifestName+"."+string(c.format))
}
func (c *Manifest) ConfigExists(fs FileSystem) bool {
	exists := true
	if _, err := fs.Stat(c.ConfigPath()); errors.Is(err, os.ErrNotExist) {
		exists = false
	}
	return exists
}

func NewJSONConfig(c Manifest) Manifest {
	return Manifest{
		format:   ConfigJson,
		TenantID: c.TenantID,
		Project:  c.Project,
	}
}

func NewYAMLConfig(c Manifest) Manifest {
	return Manifest{
		format:   ConfigYaml,
		TenantID: c.TenantID,
		Project:  c.Project,
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

const (
	ProtomokDir                        = "protomok"
	FunctionsDir                       = "functions"
	DeploymentManifestName             = "deployment"
	DataDir                            = "data"
	ErrAlreadyExists       ConfigError = "local protomok project may already exist"
)

func GetDeploymentManifest(opts ...Option) (*os.File, error) {
	// Load options, applying defaults
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	// get the filesystem. (Mocked during tests)
	fs := options.FileSystem
	dir, err := fs.Getwd()
	if err != nil {
		return nil, err
	}

	// check if we have a yaml or json deployment manifest
	jsonmf := NewJSONConfig(Manifest{})
	yamlmf := NewYAMLConfig(Manifest{})
	jsonmf.rootDir = path.Join(dir, ProtomokDir)
	yamlmf.rootDir = path.Join(dir, ProtomokDir)
	jsonOk := jsonmf.ConfigExists(fs)
	yamlOk := yamlmf.ConfigExists(fs)
	if !jsonOk && !yamlOk {
		return nil, errors.New("deployment manifest not found")
	}

	if jsonOk {
		return fs.Open(jsonmf.ConfigPath())
	}
	return fs.Open(yamlmf.ConfigPath())

}

// HasProject checks whether a protomok directory project
// exists in the current working directory.
// Returns the root path of the project. If the project does not exist,
// it will return an empty string and a nil error
func HasProject(opts ...Option) (string, error) {
	// Load options, applying defaults
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	// get the filesystem. (Mocked during tests)
	fs := options.FileSystem
	dir, err := fs.Getwd()
	if err != nil {
		return "", err
	}

	p := path.Join(dir, ProtomokDir)
	if _, err := fs.Stat(p); errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	return p, nil
}

// Initializes a new protomok directory. Returns an error if one already exists
func InitializeProject(c *Manifest, opts ...Option) (string, error) {
	// Load options, applying defaults
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	// get the filesystem. (Mocked during tests)
	fs := options.FileSystem
	dir, err := fs.Getwd()
	if err != nil {
		return "", err
	}

	alreadyExists := true
	if _, err := fs.Stat(path.Join(dir, ProtomokDir)); errors.Is(err, os.ErrNotExist) {
		alreadyExists = false
	}
	if alreadyExists {
		return "", ErrAlreadyExists
	}

	// set the root directory
	c.rootDir = path.Join(dir, ProtomokDir)

	return createProtomokFiles(c, fs)

}

func createProtomokFiles(c *Manifest, fs FileSystem) (string, error) {
	// create the root directory first
	if err := fs.Mkdir(c.rootDir, 0755); err != nil {
		return "", err
	}

	// create the functions directory
	if err := fs.Mkdir(c.FunctionDir(), 0755); err != nil {
		return "", err
	}
	// create the data directory where mocks are stored
	if err := fs.Mkdir(c.DataDir(), 0755); err != nil {
		return "", err
	}
	// create the deployment manifest file
	data, err := marshal(c)
	if err != nil {
		return "", err
	}
	if err := fs.WriteFile(c.ConfigPath(), data, 0755); err != nil {
		return "", err
	}

	return c.rootDir, nil
}

func marshal(c *Manifest) ([]byte, error) {
	if c.format == ConfigJson {
		return json.Marshal(c)
	}
	return yaml.Marshal(c)
}
