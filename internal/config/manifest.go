package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/protomoks/pmok/internal/utils/constants"
	"gopkg.in/yaml.v3"
)

type ConfigFormat string

var (
	ConfigYaml             ConfigFormat = "yaml"
	ConfigJson             ConfigFormat = "json"
	ProtomokDir                         = "protomok"
	FunctionsDir                        = filepath.Join(ProtomokDir, "functions")
	DeploymentManifestJson              = filepath.Join(ProtomokDir, "pmok.json")
	DeploymentManifestYaml              = filepath.Join(ProtomokDir, "pmok.yaml")
	ErrAlreadyExists                    = errors.New("local protomok project may already exist")
)

type Project struct {
	Name string `json:"name" yaml:"name"`
}

type ManifestConfig struct {
	rootDir   string
	format    ConfigFormat
	Version   string         `json:"version" yaml:"version"`
	Project   Project        `json:"project" yaml:"project"`
	Functions FunctionConfig `json:"functions" yaml:"functions"`
}

// initialize a default Manifest
var Manifest = &ManifestConfig{
	format:  ConfigYaml,
	Version: constants.DefaultManifestVersion,
}

func (c *ManifestConfig) ConfigPath() string {
	if c.format == ConfigYaml {
		return filepath.Join(c.rootDir, DeploymentManifestYaml)
	}
	return filepath.Join(c.rootDir, DeploymentManifestJson)

}

func (c *ManifestConfig) Copy() *ManifestConfig {
	var m ManifestConfig
	m.rootDir = c.rootDir
	m.format = c.format
	m.Version = c.Version
	m.Project = c.Project
	m.Functions = c.Functions

	return &m
}

func (c ManifestConfig) String() string {
	b, err := json.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(b)
}

func NewJSONConfig(c ManifestConfig) ManifestConfig {
	return ManifestConfig{
		Version:   constants.DefaultManifestVersion,
		format:    ConfigJson,
		Project:   c.Project,
		Functions: make(FunctionConfig),
	}
}

func NewYAMLConfig(c ManifestConfig) ManifestConfig {
	return ManifestConfig{
		Version:   constants.DefaultManifestVersion,
		format:    ConfigYaml,
		Project:   c.Project,
		Functions: make(FunctionConfig),
	}
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

	p := filepath.Join(dir, ProtomokDir)
	if _, err := fs.Stat(p); errors.Is(err, os.ErrNotExist) {
		return "", nil
	}

	return p, nil
}

func ResolveProjectDir(currentDir string) (string, error) {
	start := currentDir

	for {
		target := filepath.Join(start, ProtomokDir)
		if info, err := os.Stat(target); err == nil && info.IsDir() {
			return filepath.Dir(target), nil
		}
		// move up
		parent := path.Dir(start)
		// check if we reached "/"
		if parent == start {
			return "", fmt.Errorf("unable to find protomok project directory")
		}
		start = parent
	}
}

// Initializes a new protomok directory in the current working directory.
// Returns an error if one already exists
func InitializeProject(c *ManifestConfig, opts ...Option) (string, error) {
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
	if _, err := fs.Stat(filepath.Join(dir, ProtomokDir)); errors.Is(err, os.ErrNotExist) {
		alreadyExists = false
	}
	if alreadyExists {
		return "", ErrAlreadyExists
	}

	// set the root directory
	c.rootDir = filepath.Join(dir, ProtomokDir)

	return createProtomokFiles(c, fs)

}

func createProtomokFiles(c *ManifestConfig, fs FileSystem) (string, error) {
	// create the root directory first
	if err := fs.Mkdir(c.rootDir, 0755); err != nil {
		return "", err
	}

	// create the functions directory
	if err := fs.Mkdir(FunctionsDir, 0755); err != nil {
		return "", err
	}
	// create the data directory where mocks are stored
	// if err := fs.Mkdir(c.DataDir(), 0755); err != nil {
	// 	return "", err
	// }
	// create the deployment manifest file
	data, err := marshal(c)
	if err != nil {
		return "", err
	}

	conf := DeploymentManifestYaml
	if c.format == ConfigJson {
		conf = DeploymentManifestYaml
	}
	if err := fs.WriteFile(conf, data, 0755); err != nil {
		return "", err
	}

	return c.rootDir, nil
}

func marshal(c *ManifestConfig) ([]byte, error) {
	if c.format == ConfigJson {
		return json.MarshalIndent(c, "", "\t")
	}
	return yaml.Marshal(c)
}

func unmarshal(data []byte, m *ManifestConfig, format ConfigFormat) error {
	if format == ConfigJson {
		return json.Unmarshal(data, m)
	}
	return yaml.Unmarshal(data, m)
}
