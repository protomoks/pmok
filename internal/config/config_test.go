package config_test

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/protomoks/pmok/internal/config"
	"gopkg.in/yaml.v3"
)

type MockFileSystem struct {
	StatFunc      func(string) (os.FileInfo, error)
	GetwdFunc     func() (string, error)
	MkdirFunc     func(name string, perm os.FileMode) error
	WriteFileFunc func(name string, data []byte, perm os.FileMode) error
	OpenFunc      func(name string) (*os.File, error)
}

func (m MockFileSystem) Stat(name string) (os.FileInfo, error) {
	return m.StatFunc(name)
}

func (m MockFileSystem) Getwd() (string, error) {
	return m.GetwdFunc()
}

func (m MockFileSystem) Mkdir(name string, perm os.FileMode) error {
	return m.MkdirFunc(name, perm)
}

func (m MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return m.WriteFileFunc(name, data, perm)
}
func (m MockFileSystem) Open(name string) (*os.File, error) {
	return m.OpenFunc(name)
}

func TestInitializeProject(t *testing.T) {

	mockFs := MockFileSystem{
		GetwdFunc:     func() (string, error) { return "/tmp", nil },
		StatFunc:      func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		MkdirFunc:     func(name string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
	}
	conf := config.NewYAMLConfig(config.Manifest{TenantID: "Test", Project: config.Project{Name: "Test", ID: "1234"}})
	_, err := config.InitializeProject(&conf, config.WithFileSystem(mockFs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeProjectAlreadyExist(t *testing.T) {
	mockFs := MockFileSystem{
		GetwdFunc:     func() (string, error) { return "/tmp", nil },
		StatFunc:      func(name string) (os.FileInfo, error) { return nil, nil },
		MkdirFunc:     func(name string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
	}

	conf := config.NewYAMLConfig(config.Manifest{TenantID: "Test", Project: config.Project{Name: "Test", ID: "1234"}})
	_, err := config.InitializeProject(&conf, config.WithFileSystem(mockFs))
	if !errors.Is(err, config.ErrAlreadyExists) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitializeProjectFiles_YAML(t *testing.T) {
	dirs := make(map[string]int)
	confFile := make(map[string]config.Manifest)

	mockFs := MockFileSystem{
		GetwdFunc: func() (string, error) { return "/tmp", nil },
		StatFunc:  func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		MkdirFunc: func(name string, perm os.FileMode) error {
			dirs[name] = 1
			return nil
		},
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error {
			ext := path.Ext(name)
			var c config.Manifest
			if ext == string(config.ConfigJson) {
				json.Unmarshal(data, &c)
				confFile[name] = c
			} else {
				yaml.Unmarshal(data, &c)
				confFile[name] = c
			}
			return nil
		},
	}
	conf := config.NewYAMLConfig(config.Manifest{TenantID: "Test Tenant", Project: config.Project{Name: "Test", ID: "1234"}})
	_, err := config.InitializeProject(&conf, config.WithFileSystem(mockFs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := dirs["/tmp/protomok"]; !ok {
		t.Fatalf("expected the protomok directory to be created")
	}
	if _, ok := dirs["/tmp/protomok/functions"]; !ok {
		t.Fatalf("exptected the protomok/functions directory to be created")
	}

	if _, ok := confFile["/tmp/protomok/deployment.yaml"]; !ok {
		t.Fatal("expected the protomok/deployment.yaml file to be created")
	}

	c := confFile["/tmp/protomok/deployment.yaml"]
	if c.Project.Name != conf.Project.Name {
		t.Fatalf("conf does not match. Expected project name %s, but got %s", conf.Project.Name, c.Project.Name)
	}
	if c.TenantID != conf.TenantID {
		t.Fatalf("conf does not match. Expected project name %s, but got %s", conf.TenantID, c.TenantID)
	}

}

func TestInitializeProjectFiles_JSON(t *testing.T) {
	dirs := make(map[string]int)
	confFile := make(map[string]config.Manifest)

	mockFs := MockFileSystem{
		GetwdFunc: func() (string, error) { return "/tmp", nil },
		StatFunc:  func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		MkdirFunc: func(name string, perm os.FileMode) error {
			dirs[name] = 1
			return nil
		},
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error {
			ext := path.Ext(name)
			var c config.Manifest
			if ext == string(config.ConfigJson) {
				json.Unmarshal(data, &c)
				confFile[name] = c
			} else {
				yaml.Unmarshal(data, &c)
				confFile[name] = c
			}
			return nil
		},
	}
	conf := config.NewJSONConfig(config.Manifest{TenantID: "Test Tenant", Project: config.Project{Name: "Test Project", ID: "1234"}})
	_, err := config.InitializeProject(&conf, config.WithFileSystem(mockFs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := dirs["/tmp/protomok"]; !ok {
		t.Fatalf("expected the protomok directory to be created")
	}
	if _, ok := dirs["/tmp/protomok/functions"]; !ok {
		t.Fatalf("exptected the protomok/functions directory to be created")
	}

	if _, ok := confFile["/tmp/protomok/deployment.json"]; !ok {
		t.Fatal("expected the protomok/deployment.json file to be created")
	}

	c := confFile["/tmp/protomok/deployment.json"]
	if c.Project.Name != conf.Project.Name {
		t.Fatalf("conf does not match. Expected project name %s, but got %s", conf.Project.Name, c.Project.Name)
	}
	if c.TenantID != conf.TenantID {
		t.Fatalf("conf does not match. Expected project name %s, but got %s", conf.TenantID, c.TenantID)
	}

}

func TestGetDeploymentManifestYaml(t *testing.T) {
	mockFs := MockFileSystem{
		GetwdFunc: func() (string, error) { return "/tmp", nil },
		StatFunc: func(name string) (os.FileInfo, error) {
			// pretend like we have a yaml config
			if strings.Contains(name, ".yaml") {
				return nil, nil
			}
			return nil, os.ErrNotExist
		},
		MkdirFunc:     func(name string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
		OpenFunc:      func(name string) (*os.File, error) { return os.CreateTemp("", path.Base(name)) },
	}

	file, err := config.GetDeploymentManifest(config.WithFileSystem(mockFs))
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	if file == nil {
		t.Fatalf("expected file to be defined, but got nil")
	}
	defer os.Remove(file.Name())
}

func TestGetDeploymentManifestJson(t *testing.T) {
	mockFs := MockFileSystem{
		GetwdFunc: func() (string, error) { return "/tmp", nil },
		StatFunc: func(name string) (os.FileInfo, error) {
			// pretend like we have a json config
			if strings.Contains(name, ".json") {
				return nil, nil
			}
			return nil, os.ErrNotExist
		},
		MkdirFunc:     func(name string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
		OpenFunc:      func(name string) (*os.File, error) { return os.CreateTemp("", path.Base(name)) },
	}

	file, err := config.GetDeploymentManifest(config.WithFileSystem(mockFs))
	if err != nil {
		t.Fatalf("expected nil error, but got %s", err)
	}

	if file == nil {
		t.Fatalf("expected file to be defined, but got nil")
	}
	defer os.Remove(file.Name())
}

func TestGetDeploymentManifest_NotExist(t *testing.T) {
	mockFs := MockFileSystem{
		GetwdFunc: func() (string, error) { return "/tmp", nil },
		StatFunc: func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		},
		MkdirFunc:     func(name string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
		OpenFunc:      func(name string) (*os.File, error) { return os.CreateTemp("", path.Base(name)) },
	}

	file, err := config.GetDeploymentManifest(config.WithFileSystem(mockFs))
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}
	msg := err.Error()
	if msg != "deployment manifest not found" {
		t.Fatalf("expected error %s, but got %s", "deployment manifest not found", msg)
	}

	if file != nil {
		t.Fatalf("expected file to be nil")
	}
}
